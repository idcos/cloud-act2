#!/usr/bin/env python
#coding=utf-8

import requests
import os
import sys
import subprocess
import logging
import argparse
import errno
import json
import time
from functools import wraps
from shlex import split

__version__ = '0.1.0'
SUPPORT_SERVICE_TYPES = ('salt', 'puppet', 'openssh')


def str2bool(v):
    if v.lower() in ('yes', 'true', 't', 'y', '1'):
        return True
    elif v.lower() in ('no', 'false', 'f', 'n', '0'):
        return False
    else:
        raise argparse.ArgumentTypeError('Boolean value expected.')

def is_valid_server_addr(server_addr):
    if not server_addr:
        return False
    if server_addr.startswith('http://') or server_addr.startswith('https://'):
        return True
    else:
        return False


parser = argparse.ArgumentParser(description='salt-heartbeat')
parser.add_argument('-d', '--debug', dest='debug', type=str2bool, nargs='?',
                    const=True, help='debug, default: false')
parser.add_argument('--log-level', dest='log_level',
                    help='log level, debug, info, warning, error, default: info')
parser.add_argument('--log-file', dest='log_file', 
                help='log file, if not set, using stdout instead')
parser.add_argument('--salt-server', dest='salt_server', 
                help='salt server address, default: http://localhost:8001')
parser.add_argument('--salt-user', dest='salt_user', 
                help='salt user, default: salt-api')
parser.add_argument('--salt-password', dest='salt_password', 
                help='salt password')
parser.add_argument('--act2-server', dest='act2_server', 
                help='cloud act2 server address, eg: http://192.168.1.9')
parser.add_argument('--srv-type', dest='service_type', 
                help='service type, eg salt|puppet|openssh')                
parser.add_argument('--idc', dest='idc', 
                help='idc information')
parser.add_argument('-v', '--version', action='version',
                    version='%(prog)s {version}'.format(version=__version__))                


parser.set_defaults(log_level='info',
    salt_host='localhost', salt_port='8001',
    salt_user='salt-api',
    debug=False,
    service_type='salt',
)

args = parser.parse_args()


# init logger
log_level_map = {
    'debug': logging.DEBUG,
    'info': logging.INFO,
    'warning': logging.WARNING,
    'error': logging.ERROR,
}

log_level = log_level_map.get(args.log_level)
log_file = args.log_file


# logging.Formatter
logger = logging.getLogger('heartbeat')
logger.setLevel(log_level)
formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
if log_file:
    fh = logging.FileHandler(log_file)
    fh.setLevel(log_level)
    fh.setFormatter(formatter)
    logger.addHandler(fh)
else:
    ch = logging.StreamHandler()
    ch.setLevel(log_level)
    ch.setFormatter(formatter)
    logger.addHandler(ch)


debug = args.debug


# check act2 server must valid
act2_server = args.act2_server
if debug and not act2_server:
    act2_server = 'http://localhost:6868'

act2_server = act2_server.rstrip('/')
if not is_valid_server_addr(act2_server):
    print('act2 server not valid')
    parser.print_usage()
    sys.exit(1)


idc = args.idc
if not idc:
    print('must given idc information')
    parser.print_usage()
    sys.exit(1)


# salt information
salt_server = args.salt_server
salt_user = args.salt_user
salt_password = args.salt_password
if salt_user == 'salt-api' and not salt_password:
    salt_password = 'UiuQWQuqmVsrQ5KJpi8pKQ=='


salt_server = salt_server.rstrip('/')
if not is_valid_server_addr(salt_server):
    print('salt server not valid')
    parser.print_usage()
    sys.exit(1)

service_type = args.service_type

if service_type not in SUPPORT_SERVICE_TYPES:
    print('service type is invalid')
    parser.print_usage()
    sys.exit(1)

## 注意，mac下是没有对应的`/sbin/ip`执行文件，所以该脚本只能在linux下运行
def get_remote_ip():
    p1 = subprocess.Popen(split("/sbin/ip -4 route get 10.0.0.0"), stdout=subprocess.PIPE)
    p2 = subprocess.Popen(['awk', '/src/ {sub("metric.*",""); print $NF }'], stdin=p1.stdout, stdout=subprocess.PIPE)
    remote_ip =  p2.stdout.read().strip()
    return remote_ip


class SaltClient(object):
    headers = {
        'Accept': 'application/json',
        'Content-type': 'application/json',
        'User-Agent': 'idcos cloudact2',
    }

    def __init__(self, server, user, password):
        self.server = server
        self.user = user
        self.password = password

    def get_token(self):
        login_url = '%s/login' % self.server
        headers = self.headers.copy()
        data = {
            'username': self.user,
            'password': self.password,
            'eauth': 'pam',
        }
        logger.debug('login_url %s, data %s, headers %s', login_url, data, headers)
        resp = requests.post(login_url, json=data, headers=headers)
        if resp.status_code == 200:
        # TODO: 需要处理异常状态
            resp_data = resp.json()
            token = resp_data['return'][0]['token']
            return token
        else:
            logger.info('get token error, status code %s, content %s', resp.status_code, resp.content)
            return ''

    def get_application(self):
        token = self.get_token()
        if not token:
            return {}
        headers = self.headers.copy()
        headers['X-Auth-Token'] = token
        stats_url = '%s/stats' % self.server
        resp = requests.get(stats_url, headers=headers)
        stats = resp.json()
        cherry = stats['CherryPy Applications']
        return cherry

    def get_bind_addr(self):
        application = self.get_application()
        if not application:
            logger.info("get an empty application")
            return ''
        bind_address = application['Bind Address']
        addr = bind_address.lstrip('(').rstrip(')').split(',')
        host = addr[0].strip("'")
        if host == '0.0.0.0':
            host = get_remote_ip()
        return host

    def status(self):
        '''运行状态返回'''
        application = self.get_application()
        if not application:
            logger.info("get an empty application")
            return ''
        #  u'Bind Address': u"('0.0.0.0', 8000)",
        uptime = application.get('Uptime', None)
        return uptime is not None

    def minions(self):
        token = self.get_token()
        if not token:
            return []
        headers = self.headers.copy()
        headers['X-Auth-Token'] = token
        minion_url = '%s/minions' % self.server
        resp = requests.get(minion_url, headers=headers)
        if resp.status_code == 200:
            resp_data = resp.json()
            minions = resp_data.pop('return')
            return minions
        else:
            logger.info('get minion error, status code %s, content %s', resp.status_code, resp.content)
            return []


class FileLockError(StandardError):
    pass


class FileLock(object):

    def __init__(self, filename, timeout=10, delay=0.05):
        if timeout is not None and delay is None:
            raise ValueError("If timeout is not None, then delay must not be None.")
        self.is_locked = False
        self.filename = filename
        self.fd = None
        self.timeout = timeout
        self.delay = delay

    def accquire(self):
        start_time = time.time()
        while True:
            try:
                self.fd = os.open(self.filename, os.O_CREAT|os.O_EXCL|os.O_RDWR)
                self.is_locked = True
            except OSError as e:
                if e.errno != errno.EEXIST:
                    raise
                if self.timeout is None:
                    raise FileLockError("Could not acquire lock on {}".format(self.filename))
                if (time.time() - start_time) >= self.timeout:
                    raise FileLockError("Timeout occured.")
    
    def release(self):
        if self.is_locked:
            os.close(self.fd)
            os.unlink(self.filename)
            self.is_locked = False

    def __enter__(self):
        if not self.is_locked:
            try:
                self.accquire()
            except FileLockError as e:
                logger.error('file lock error %s', e)
        return self

    def __exit__(self, type, value, traceback):
        if self.is_locked:
            self.release()

    def __del__(self):
        self.release()


def singlelock(func):
    @wraps(func)
    def warpper(*args, **kwargs):
        with FileLock("/tmp/salt-heartbeat.lock"):
            return func(*args, **kwargs)
    return warpper


def get_master_info():
    master = {
        'server': salt_server,
        'options': {
            'username': salt_user,
            'password': salt_password,
        },
        'status': 'running',
        'idc': idc,
        'type': 'salt',
        'sn': get_master_sn(),
    }
    return master

def get_master_sn():
    file = '/sys/class/dmi/id/product_uuid'
    with open(file, 'r') as fp:
        sn = fp.read()
        sn.lower()
        return sn

def get_ip4_interfaces(ip4_interfaces):
    ip4 = []
    for key, value in ip4_interfaces.iteritems():
        ips = [ip for ip in value if ip not in ('127.0.0.1')]
        ip4.extend(ips)
    return ip4


def get_minion_grains():
    salt_client = SaltClient(salt_server, salt_user, salt_password)
    minions = salt_client.minions()

    new_grains = []
    for minions_object in minions:
        logger.debug('minion info %s', minions_object)
        for sn, minion in minions_object.iteritems():
            if isinstance(minion, bool):
                if not minion:
                    new_grains.append({
                        'sn': sn,
                        'status': 'stopped',
                    })
                else:
                    logger.warn('grain minion is true, set status to running')
                    new_grains.append({
                        'sn': sn,
                        'status': 'running',
                    })
            else:
                logger.info('get minion info %s', sn)
                ip4_interfaces = minion['ip4_interfaces']
                ip4_interfaces = get_ip4_interfaces(ip4_interfaces)

                new_grains.append({
                    "sn": sn,
                    "ips": ip4_interfaces,
                    "status": 'running',
                })
    return new_grains


def get_minions_info():
    minion_grains = get_minion_grains()
    logger.info('get minion grains %s', minion_grains)
    return minion_grains


# @singlelock
def heartbeat():
    minions =  get_minions_info()
    master = get_master_info()
    data = {
        'master': master,
        'minions': minions,
    }
    logger.debug('register data %s', data)
    act2_addr = '%s/register' %act2_server
    try:
        resp = requests.post(act2_addr, json=data)
        if resp.status_code == 200:
            logger.info('heartbeat to %s success', act2_server)
        else:
            logger.error('heartbeat to %s fail, status code %s, resp: %s', act2_server, resp.status_code, resp.content)
    except requests.exceptions.RequestException as e:
        logger.error('register to server error %s', e)


def main():
    if debug:
        logger.debug('salt server %s, user %s', salt_server, salt_user)

    heartbeat()


if __name__ == '__main__':
    main()