CREATE DATABASE IF NOT EXISTS `cloud-act2` CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci';
use `cloud-act2`;

CREATE TABLE act2_api_log
(
  id             VARCHAR(64)  NOT NULL
    PRIMARY KEY,
  url            VARCHAR(255) NULL
  COMMENT '接口地址',
  description    VARCHAR(255) NULL
  COMMENT '接口描述',
  type           VARCHAR(32)  NULL
  COMMENT '请求方式',
  user           VARCHAR(64)  NULL
  COMMENT '操作者',
  addr           VARCHAR(64)  NULL
  COMMENT '请求地址',
  operate_time   DATETIME     NULL
  COMMENT '操作时间',
  time_consuming INT          NULL
  COMMENT '耗时',
  request_params LONGTEXT     NULL
  COMMENT '请求参数',
  response_body  LONGTEXT     NULL
  COMMENT '结果详情'
)
  COMMENT 'api日志'
  ENGINE = InnoDB
  CHARSET = utf8mb4;

CREATE TABLE act2_host
(
  id             VARCHAR(64)            NOT NULL
  COMMENT '主键'
    PRIMARY KEY,
  idc_id         VARCHAR(64)            NOT NULL
  COMMENT '机房信息',
  entity_id      VARCHAR(64) DEFAULT '' NULL
  COMMENT '主机唯一标识，主机sn',
  add_time       DATETIME               NOT NULL
  COMMENT '添加时间',
  status         VARCHAR(64)            NULL
  COMMENT '状态',
  os_type        VARCHAR(64)            NULL
  COMMENT '系统类型，windows|linux|aix',
  proxy_id       VARCHAR(64)            NOT NULL
  COMMENT 'proxy id',
  minion_version VARCHAR(64)            NULL
  COMMENT 'salt的版本',
  CONSTRAINT entity_id
  UNIQUE (entity_id)
)
  COMMENT '主机列表'
  ENGINE = InnoDB
  CHARSET = utf8mb4;

CREATE TABLE act2_host_ip
(
  id       VARCHAR(64) NOT NULL
  COMMENT '主键'
    PRIMARY KEY,
  host_id  VARCHAR(64) NOT NULL
  COMMENT 'MINION ID',
  ip       VARCHAR(16) NULL
  COMMENT '主机IP',
  add_time DATETIME    NULL
  COMMENT '修改时间'
)
  COMMENT 'IP列表'
  ENGINE = InnoDB
  CHARSET = utf8mb4;

CREATE TABLE act2_host_result
(
  id             VARCHAR(64) NOT NULL
  COMMENT '主键'
    PRIMARY KEY,
  task_id        VARCHAR(64) NOT NULL
  COMMENT '作业执行记录id',
  host_id        VARCHAR(64) NOT NULL
  COMMENT '主机id',
  proxy_id       VARCHAR(64) NULL
  COMMENT '下发的act2_proxy',
  start_time     DATETIME    NOT NULL
  COMMENT '开始时间',
  end_time       DATETIME    NULL
  COMMENT '结束时间',
  execute_status VARCHAR(16) NOT NULL
  COMMENT '执行状态:DOING:正在执行|DONE:执行完毕|CANCELLED',
  result_status  VARCHAR(16) NULL
  COMMENT '执行结果:SUCCESS:成功|FAIL:失败|TIMEOUT',
  host_ip        VARCHAR(64) NULL
  COMMENT '主机ip',
  stdout         LONGTEXT    NULL
  COMMENT '输出结果',
  stderr         LONGTEXT    NULL
  COMMENT '输出错误',
  message        LONGTEXT    NULL
  COMMENT '调用proxy结果或proxy回调异常结果信息'
)
  COMMENT '主机执行结果表'
  ENGINE = InnoDB
  CHARSET = utf8mb4;

CREATE INDEX task_id
  ON act2_host_result (task_id);

CREATE INDEX host_id
  ON act2_host_result (host_id);

CREATE TABLE act2_idc
(
  id       VARCHAR(64) NOT NULL
  COMMENT '主键'
    PRIMARY KEY,
  name     VARCHAR(64) NOT NULL
  COMMENT '机房名称',
  add_time DATETIME    NULL
  COMMENT '修改时间',
  CONSTRAINT name
  UNIQUE (name)
)
  COMMENT '机房信息表'
  ENGINE = InnoDB
  CHARSET = utf8mb4;

CREATE TABLE act2_job_record
(
  id             VARCHAR(64)            NOT NULL
  COMMENT '主键'
    PRIMARY KEY,
  start_time     DATETIME               NOT NULL
  COMMENT '创建时间',
  end_time       DATETIME               NULL
  COMMENT '修改时间',
  execute_status VARCHAR(16)            NOT NULL
  COMMENT '执行状态:DOING:正在执行|DONE:执行完毕|CANCELLED',
  result_status  VARCHAR(16)            NULL
  COMMENT '执行结果:SUCCESS:成功|FAIL:失败|TIMEOUT',
  callback       VARCHAR(64)            NULL
  COMMENT '其他系统回调地址',
  provider       VARCHAR(64)            NULL
  COMMENT 'salt|puppet|openssh',
  pattern        VARCHAR(64)            NULL
  COMMENT '模块名称：file、script、salt.state',
  script         LONGTEXT               NULL
  COMMENT '脚本内容||文件内容',
  script_type    VARCHAR(64)            NULL
  COMMENT '脚本类型: python|shell...',
  timeout        INT                    NULL
  COMMENT '超时时间',
  parameters     LONGTEXT               NULL
  COMMENT '参数信息',
  hosts          LONGTEXT               NOT NULL,
  user           VARCHAR(64) DEFAULT '' NULL
  COMMENT '外部调用用户名',
  master_id      VARCHAR(64)            NOT NULL
  COMMENT 'master entity id',
  execute_id     VARCHAR(64)            NULL
  COMMENT '外部任务的，执行id'
)
  COMMENT '作业执行记录表'
  ENGINE = InnoDB
  CHARSET = utf8mb4;

CREATE INDEX id
  ON act2_job_record (id);

CREATE INDEX start_time
  ON act2_job_record (start_time);

CREATE INDEX execute_status
  ON act2_job_record (execute_status);

CREATE INDEX master_id
  ON act2_job_record (master_id);

CREATE TABLE act2_job_task
(
  id             VARCHAR(64) NOT NULL
  COMMENT '主键'
    PRIMARY KEY,
  start_time     DATETIME    NOT NULL
  COMMENT '创建时间',
  end_time       DATETIME    NULL
  COMMENT '修改时间',
  execute_status VARCHAR(16) NOT NULL
  COMMENT '执行状态:DOING:正在执行|DONE:执行完毕|CANCELLED',
  result_status  VARCHAR(16) NULL
  COMMENT '执行结果:SUCCESS:成功|FAIL:失败|TIMEOUT',
  record_id      VARCHAR(64) NULL
  COMMENT 'salt|puppet|openssh',
  pattern        VARCHAR(64) NULL
  COMMENT '模块名称：file、script、salt.state',
  script         LONGTEXT    NULL
  COMMENT '脚本内容||文件内容',
  params         LONGTEXT    NULL
  COMMENT '参数信息',
  options        TEXT        NULL
  COMMENT '配置参数'
)
  COMMENT '作业执行任务表'
  ENGINE = InnoDB
  CHARSET = utf8mb4;

CREATE INDEX record_id
  ON act2_job_task (record_id);

CREATE TABLE act2_job_task_proxy
(
  id             VARCHAR(64) NOT NULL
    PRIMARY KEY,
  task_id        VARCHAR(64) NULL
  COMMENT 'task的主键',
  proxy_id       VARCHAR(64) NULL
  COMMENT 'proxy的主键',
  start_time     DATETIME    NOT NULL
  COMMENT '创建时间',
  end_time       DATETIME    NULL
  COMMENT '修改时间',
  execute_status VARCHAR(16) NOT NULL
  COMMENT '执行状态:DOING:正在执行|DONE:执行完毕|CANCELLED',
  result_status  VARCHAR(16) NULL
  COMMENT '执行结果:SUCCESS:成功|FAIL:失败|TIMEOUT'
)
  COMMENT '作业执行任务proxy表'
  ENGINE = InnoDB
  CHARSET = utf8mb4;

CREATE TABLE act2_proxy
(
  id         VARCHAR(64)  NOT NULL
  COMMENT '主键'
    PRIMARY KEY,
  last_time  DATETIME     NOT NULL
  COMMENT '上一次上报时间',
  twice_time DATETIME     NULL
  COMMENT '上上次上报时间',
  idc_id     VARCHAR(64)  NOT NULL
  COMMENT '机房信息',
  server     VARCHAR(128) NULL
  COMMENT '服务地址，如http://192.168.1.19:8000/',
  type       VARCHAR(64)  NULL
  COMMENT '类型 puppet|salt...',
  status     VARCHAR(64)  NULL
  COMMENT '状态',
  options    VARCHAR(64)  NULL
  COMMENT '可选数据'
)
  COMMENT 'act2 proxy'
  ENGINE = InnoDB
  CHARSET = utf8mb4;