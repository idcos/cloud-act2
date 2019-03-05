DROP DATABASE IF EXISTS `cloud-act2`;
CREATE DATABASE `cloud-act2`;

USE `cloud-act2`;


DROP TABLE IF EXISTS `act2_host_result`;
CREATE TABLE `act2_host_result` (
  `id` varchar(64) NOT NULL COMMENT '主键',
  `job_record_id` varchar(64) NOT NULL COMMENT '作业执行记录id',
  `host_id` varchar(64) NOT NULL COMMENT '主机id',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime DEFAULT NULL COMMENT '结束时间',
  `execute_status` varchar(16) NOT NULL COMMENT '执行状态:DOING:正在执行|DONE:执行完毕|CANCELLED',
  `result_status` varchar(16) DEFAULT NULL COMMENT '执行结果:SUCCESS:成功|FAIL:失败|TIMEOUT',
  `host_ip` varchar(64) DEFAULT NULL COMMENT '主机ip',
  `stdout` longtext COMMENT '输出结果',
  `stderr` longtext COMMENT '输出错误',
  `message` varchar(256) DEFAULT NULL COMMENT '调用proxy结果或proxy回调异常结果信息',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主机执行结果表';


DROP TABLE IF EXISTS `act2_job_record`;
CREATE TABLE `act2_job_record` (
  `id` varchar(64) NOT NULL COMMENT '主键',
  `start_time` datetime NOT NULL COMMENT '创建时间',
  `end_time` datetime DEFAULT NULL COMMENT '修改时间',
  `execute_status` varchar(16) NOT NULL COMMENT '执行状态:DOING:正在执行|DONE:执行完毕|CANCELLED',
  `result_status` varchar(16) DEFAULT NULL COMMENT '执行结果:SUCCESS:成功|FAIL:失败|TIMEOUT',
  `callback` varchar(64) DEFAULT NULL COMMENT '其他系统回调地址',
  `proxy_id` varchar(64) DEFAULT NULL COMMENT '下发的act2_proxy',
  `provider` varchar(64) DEFAULT NULL COMMENT 'salt|puppet|openssh',
  `module_name` varchar(64) DEFAULT NULL COMMENT '模块名称：file、script、salt.state',
  `script` longtext COMMENT '脚本内容||文件内容',
  `script_type` varchar(64) DEFAULT NULL COMMENT '脚本类型: python|shell...',
  `timeout` int(11) DEFAULT NULL COMMENT '超时时间',
  `parameters` longtext COMMENT '参数信息',
  `hosts` longtext NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作业执行记录表';


DROP TABLE IF EXISTS `act2_idc`;
CREATE TABLE `act2_idc` (
  `id` varchar(64) NOT NULL COMMENT '主键',
  `name` varchar(64) NOT NULL COMMENT '机房名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机房信息表';

DROP TABLE IF EXISTS `act2_proxy`;
CREATE TABLE `act2_proxy` (
  `id` varchar(64) NOT NULL COMMENT '主键',
  `last_time` datetime NOT NULL COMMENT '上一次上报时间',
  `twice_time` datetime DEFAULT NULL COMMENT '上上次上报时间',
  `idc_id` varchar(64) NOT NULL COMMENT '机房信息',
  `server` varchar(128) DEFAULT NULL COMMENT '服务地址，如http://192.168.1.19:8000/',
  `type` varchar(64) DEFAULT NULL COMMENT '类型 puppet|salt...',
  `status` varchar(64) DEFAULT NULL COMMENT '状态',
  `options` varchar(64) DEFAULT NULL COMMENT '可选数据',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='act2 proxy';


DROP TABLE IF EXISTS `act2_host`;
CREATE TABLE `act2_host` (
  `id` varchar(64) NOT NULL COMMENT '主键',
  `idc_id` varchar(64) NOT NULL COMMENT '机房信息',
  `entity_id` varchar(64) default "" COMMENT '主机唯一标识，主机sn',
  `add_time` datetime NOT NULL COMMENT '添加时间',
  `status` varchar(64) DEFAULT NULL COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主机列表';


DROP TABLE IF EXISTS `act2_host_ip`;
CREATE TABLE `act2_host_ip` (
  `id` varchar(64) NOT NULL COMMENT '主键',
  `host_id` varchar(64) NOT NULL COMMENT 'MINION ID',
  `ip` varchar (16) DEFAULT NULL COMMENT '主机IP',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IP列表';

