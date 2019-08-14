package accumulate

import (
	"fmt"
	"time"
	"log"
	"strconv"
	"strings"
	//"errors"
	"encoding/json"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	// "github.com/DouwaIO/hairtail/src/schema"
	"github.com/DouwaIO/hairtail/src/model"
	// "github.com/DouwaIO/hairtail/src/store/datastore"
	"github.com/DouwaIO/hairtail/src/utils"
	"github.com/DouwaIO/hairtail/src/task"
)


type Plugin struct {
    loop_num int
    end int64
}

// func Accumulate(data []byte, params map[string]interface{}) {
func (p *Plugin) Run(params *task.Params) (*task.Result, error) {
	log.Println("Accumulate")

	start := time.Now().Unix()
	log.Println("accumulate start is ", start)
	start += 1

	var list_data []interface{}
	err := json.Unmarshal(params.Data, &list_data)
	if err != nil {
		log.Printf("%s", err)
		//return err
	}

	setting_map := params.Settings["map"]
	source := params.Settings["source"]
	target := params.Settings["target"]
	compute := params.Settings["compute"]
	ignore := params.Settings["ignore"]

	map_map := make(map[string]string)

	for i := 0; i < len(setting_map.([]interface{})); i++ {
		d := strings.Split(setting_map.([]interface{})[i].(string), "=")
		map_map[d[0]] = d[1]
	}

	db, err := gorm.Open("postgres", "host=47.110.154.127 port=30011 user=postgres dbname=postgres sslmode=disable password=huansi@2017")
	if err != nil {
		log.Printf("%s", err)
		//log.Printf(err)
		//return err
	}

	tx := db.Begin()

	for i := 0; i < len(list_data); i++ {

		// field_text := ""
		// field_value := ""
		// for key := range map_map{
		// 	field_text += fmt.Sprintf(" %s text,", map_map[key])
		// 	// field_value += fmt.Sprintf(" %s = '%s',", key,map_map[key])
		// }
		// field_text = strings.TrimRight(field_text,",")
		// field_value = strings.TrimRight(field_value,",")

		field_text := ""
		field_value := ""

		for key := range map_map {
			// fmt.Println(reflect.TypeOf(list_data[i][key]))
			val_string := ""

			val := list_data[i].(map[string]interface{})[key]

			switch val.(type) {
			case string:
				val_string = val.(string)
				// fmt.println("这是一个string类型的变量")
			case int64:
				val_string = strconv.FormatInt(val.(int64), 10)
			case float32:
				val_string = fmt.Sprintf("%g", val.(float64))
				// fmt.Print(6666)
			case float64:
				val_string = fmt.Sprintf("%g", val.(float64))
			}

			field_text += fmt.Sprintf(" %s text,", map_map[key])
			field_value += fmt.Sprintf(" o.%s = '%s' and ", map_map[key], val_string)

		}
		field_text = strings.TrimRight(field_text, ",")
		field_value = strings.TrimRight(field_value, "and ")
		// 判断数据库是否存在
		sql_str := fmt.Sprintf(`SELECT "id","name","data" FROM (
			SELECT "id","name","data" FROM remote_data as d, jsonb_to_record(d.data) o (%s) WHERE %s
		) as dd`, field_text, field_value)

		row := tx.Raw(sql_str).Row()
		var result model.RemoteData
		db_data_map := make(map[string]interface{})

		err = row.Scan(&result.ID, &result.Name, &result.Data)
		// if err != nil{
		// 	// tx.Rollback()
		// 	log.Printf("%s,rollback", err)

		// 	// return
		// }
		is_exist := false
		if result.ID != "" {
			is_exist = true
			_ = json.Unmarshal(result.Data, &db_data_map)
		}

		// 判断数据是否map匹配的条件

		if is_exist == true { //如果这条数据匹配到，并且数据库数据存在则相加
			// var target_val float64

			db_data_target_val := db_data_map[target.(string)]

			source_val := list_data[i].(map[string]interface{})[source.(string)]

			if db_data_target_val == nil || source_val == nil {
				log.Printf("%s", err)
				//return errors.New("target_val或source_value不存在")
			}

			if compute != "_" { //如果操作符不是-
				source_val = source_val.(float64)
			} else {
				source_val = source_val.(float64) * (-1)
			}

			db_data_target_val = db_data_target_val.(float64) + source_val.(float64)
			// db_data_target_val = db_data_target_val + target_val //给数据库的值重新复制
			db_data_map[target.(string)] = db_data_target_val
			// fmt.Println(db_data_map)

			byte_data, _ := json.Marshal(db_data_map)

			err := tx.Model(&result).Update("data", byte_data).Error
			if err != nil {
				tx.Rollback()
				log.Printf("%s,rollback", err)
				//return err
				return nil, err
			}

		} else {
			if ignore == false {
				source_val := list_data[i].(map[string]interface{})[source.(string)]
				//如果数据库中不存在，但是条件符合 则进行insert
				if compute != "-" { //如果操作符不是-
					source_val = source_val.(float64)
				} else {
					source_val = source_val.(float64) * (-1)
				}
				list_data[i].(map[string]interface{})[source.(string)] = source_val

				byte_data, _ := json.Marshal(list_data[i].(map[string]interface{}))
				var result model.RemoteData
				gen_id := utils.GeneratorId()
				result.Data = byte_data
				result.ID = gen_id
				err := tx.Create(&result).Error
				if err != nil {
					tx.Rollback()
					log.Printf("%s, rollback", err)
					return nil, err
					//return err
				}
			}

		}
	}

	p.loop_num += 1
	fmt.Println("loop_num is ", p.loop_num)

	p.end = time.Now().Unix()
	fmt.Println("end is ", p.end)

	tx.Commit()
	err = db.Close()
	if err != nil {
		log.Println(err)
	}

	return nil, nil
}