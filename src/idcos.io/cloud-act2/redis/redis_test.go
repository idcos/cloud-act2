//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package redis

//
//func TestSendPublish(t *testing.T) {
//	config.Conf = &config.Config{
//		Redis: config.RedisConfig{
//			Server: "127.0.0.1",
//			Port:   6379,
//		},
//	}
//
//	err := Load()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	conn := GetConn()
//	defer conn.Close()
//	_, err = conn.Do("PUBLISH", "pub1", `{"stdout": "xxxx". "stderr": "xxxx", "now": now}`)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for {
//		c := GetConn()
//
//		psc := redis.PubSubConn{Conn: c}
//		psc.Subscribe("aa")
//
//		msg := psc.Receive()
//		if msg != nil {
//			fmt.Println(msg)
//		}
//
//	}
//}
//
//func TestGetSaltResult(t *testing.T) {
//	config.Conf = &config.Config{
//		Redis: config.RedisConfig{
//			Server: "192.168.1.17",
//			Port:   6379,
//		},
//	}
//
//	err := Load()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	conn := GetConn()
//	defer conn.Close()
//	res, err := redis.Values(conn.Do("HGETALL", "ret:20181031104103607346"))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for len(res) > 0 {
//		var entityID string
//		var data string
//		res, err = redis.Scan(res, &entityID, &data)
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		var retMap interface{}
//		err = json.Unmarshal([]byte(data), &retMap)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//}
