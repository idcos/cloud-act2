#!/usr/bin/env python
#coding=utf-8

import logging
import os
import sys
import subprocess

logger = logging.getLogger('idcos_grains')

def _command(cmd):
    p = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    stdout, _ = p.communicate()

    lines = stdout.split('\r\n')
    lines = [line.strip() for line in lines if line.strip()]
    if len(lines) == 0: return ''

    start = lines[0]
    if start == 'UUID':
        # in windows
        return lines[1].upper()

    return start.upper()

def idcos_grains():
    platform = sys.platform.lower()
    if platform in ['win32', 'win64', 'windows']:
        cmd = ['wmic', 'csproduct', 'get', 'uuid']
    elif platform in ['linux', 'linux2']:
        cmd = ['cat', '/sys/class/dmi/id/product_uuid']
    elif platform in ['aix']:
        cmd =  ['uname', '-f']
    else:
        cmd = ['cat', '/sys/class/dmi/id/product_uuid']


    stdout = _command(cmd)
    grains = {
        'idcos_system_id': stdout,
    }

    return grains


if __name__ == '__main__':
    print(idcos_grains())
