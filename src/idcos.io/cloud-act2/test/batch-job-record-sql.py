#!/usr/bin/env python
#coding=utf-8

# 批量生成act2_job_record的sql，需要支持
# id为varchar(64)的方式
# id为binary(16)的方式
# 参考：https://www.percona.com/blog/2014/12/19/store-uuid-optimized-way/
# https://blog.csdn.net/JavaReact/article/details/78854283

# act2_job_record表的顺序如下：
'''
`id` varchar(64) NOT NULL COMMENT '主键',
  `start_time` datetime NOT NULL COMMENT '创建时间',
  `end_time` datetime DEFAULT NULL COMMENT '修改时间',
  `execute_status` varchar(16) NOT NULL COMMENT '执行状态:DOING:正在执行|DONE:执行完毕|CANCELLED',
  `result_status` varchar(16) DEFAULT NULL COMMENT '执行结果:SUCCESS:成功|FAIL:失败|TIMEOUT',
  `callback` varchar(64) DEFAULT NULL COMMENT '其他系统回调地址',
  `provider` varchar(64) DEFAULT NULL COMMENT 'salt|puppet|openssh',
  `pattern` varchar(64) DEFAULT NULL COMMENT '模块名称：file、script、salt.state',
  `script` longtext COMMENT '脚本内容||文件内容',
  `script_type` varchar(64) DEFAULT NULL COMMENT '脚本类型: python|shell...',
  `timeout` int(11) DEFAULT NULL COMMENT '超时时间',
  `parameters` longtext COMMENT '参数信息',
  `hosts` longtext NOT NULL,
  `master_id` varchar(64) DEFAULT '' COMMENT 'master的机器标识',
  `user` varchar(64) DEFAULT '' COMMENT '外部调用用户名',
  `execute_id` varchar(64) DEFAULT '' COMMENT '外部任务的，执行id',
'''


import uuid
import binascii
import argparse
import datetime
import random
import string
import csv
import json


def make_binary_uuid():
    uuid = make_uuid()
    new_uuid = uuid[15:4] + uuid[10:4] + uuid[1:8] + uuid[20:4] + uuid[25:]
    return binascii.a2b_hex(new_uuid)


def make_uuid():
    return str(uuid.uuid4())


def random_year():
    return random.choice([2017, 2018])

def random_month():
    return random.choice([month for month in xrange(1,12)])

def random_day(max_day):
    return random.randint(1, max_day)

def random_hour():
    return random.randint(0,23)

def random_miniute():
    return random.randint(0,59)

random_second = random_miniute

def random_str(length):
    return ''.join([random.choice(string.printable) for i in xrange(length)])


def random_execute_status():
    return random.choice(['DOING', 'DONE', 'CANCELLED'])

def random_result_status():
    return random.choice(['SUCCESS', 'FAIL', 'TIMEOUT'])

def random_provider():
    return random.choice(['salt', 'ssh', 'puppet'])

def random_pattern():
    return random.choice(['file', 'script', 'sls'])


def random_file():
    return random_str(random.randint(0, 20000))

def random_bash_script():
    return random.choice([
'''#/bin/bash
echo "hello world"
''',
'''#/bin/bash
echo '$@'
'''
    ])


def random_python_script():
    return random.choice([
'''#/usr/bin/env python
print('hello world')
''',
'''#/usr/bin/env python
import os
print (os.argv)
'''
    ])


def random_script(script_type):
    if script_type == 'bash':
        return random_bash_script()
    elif script_type == 'python':
        return random_python_script()
    else:
        return random_bash_script()


def random_script_type():
    return random.choice(['bash', 'python'])


def random_timeout():
    return random.randint(20,5000)

def random_parameters():
    return random_str(random.randint(0,30))

def random_hosts():
    hosts = ['192.168.1.%s' %i for i in xrange(2,254)]    
    count = random.randint(1,len(hosts))
    return json.dumps(list(set([random.choice(hosts) for i in xrange(count)])))

master_ids = [
    str(uuid.uuid4()),
    str(uuid.uuid4()),
    str(uuid.uuid4()),
    str(uuid.uuid4()),
]

def random_master_id():
    return random.choice(master_ids)


def random_user():
    return random.choice(['root', 'admin', 'ganyu', 'xuyuan', 'wenhe'])

def random_execute_id():
    return str(uuid.uuid4())


def random_callback_url():
    return random.choice([
        'http://192.168.1.17:6868/api/v1/callback/test',
        'http://10.0.0.124:6868/api/v1/callback/test',
    ])


class Maker(object):

    def __init__(self, id_make):
        self.id_make = id_make

    def _get_time(self):
        start = datetime.datetime(random_year(), random_month(), random_day(28), random_hour(), random_miniute(), random_second())
        delta = datetime.timedelta(days=random.randint(0,2), hours=random_hour(), minutes=random_miniute(), seconds=random_second())
        end = start + delta
        
        return start, end

    def random(self):
        start_time, end_time = self._get_time()
        pattern = random_pattern()
        script_type = random_script_type()
        script = random_script(script_type)

        return [
            self.id_make(),
            start_time,
            end_time,
            random_execute_status(),
            random_result_status(),
            random_callback_url(),
            random_provider(),
            pattern, 
            script_type,
            script,
            random_timeout(),
            random_parameters(),
            random_hosts(),
            random_master_id(),
            random_user(),
            random_execute_id(),
        ]

def get_maker(binary):
    if binary:
        return Maker(make_binary_uuid)
    else:
        return Maker(make_uuid)


parser = argparse.ArgumentParser(description='batch job record sql')
parser.add_argument('-b', '--binary', dest='binary', help='binary uuid', default=False)
parser.add_argument('-t', '--total', dest='total', help='total count', default=1000000)
parser.add_argument('-c', '--csv', dest='csv', help='csv file', default='/tmp/batch_job_record.csv')
result = parser.parse_args()

total = result.total
binary = result.binary
csv_filepath = result.csv

maker = get_maker(binary=binary)


headers = [
'id','start_time','end_time','execute_status','result_status','callback','provider','pattern','script','script_type','timeout','parameters','hosts','master_id','user','execute_id'
]

with open(csv_filepath, 'wb') as csvfile:
    writer = csv.writer(csvfile, delimiter=',', quotechar='"')
    
    for i in xrange(total):
        v = maker.random()
        writer.writerow(v)

print('''mysql -uroot -p LOAD DATA INFILE '/tmp/batch_job_record.csv' INTO TABLE cloud-act2.act2_job_record FIELDS  TERMINATED BY   ','  LINES TERMINATED BY '\n' ''')