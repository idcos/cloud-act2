
install:
	[[ -f /usr/lib/python2.7/site-packages/salt/minion.py ]] && cp -p /usr/lib/python2.7/site-packages/salt/minion.py /usr/lib/python2.7/site-packages/salt/minion.py.bak
	[[ -f /usr/lib/python2.7/site-packages/salt/modules/cmdmod.py ]] && cp -p /usr/lib/python2.7/site-packages/salt/modules/cmdmod.py /usr/lib/python2.7/site-packages/salt/modules/cmdmod.py.bak
	[[ -f /usr/lib/python2.7/site-packages/salt/utils/vt.py ]] && cp -p /usr/lib/python2.7/site-packages/salt/utils/vt.py /usr/lib/python2.7/site-packages/salt/utils/vt.py.bak
	cat minion.py > /usr/lib/python2.7/site-packages/salt/minion.py
	cat cmdmod.py > /usr/lib/python2.7/site-packages/salt/modules/cmdmod.py
	cat vt.py > /usr/lib/python2.7/site-packages/salt/utils/vt.py


uninstall:
	cp -p /usr/lib/python2.7/site-packages/salt/minion.py.bak /usr/lib/python2.7/site-packages/salt/minion.py
	cp -p /usr/lib/python2.7/site-packages/salt/modules/cmdmod.py.bak /usr/lib/python2.7/site-packages/salt/modules/cmdmod.py
	cp -p /usr/lib/python2.7/site-packages/salt/utils/vt.py.bak /usr/lib/python2.7/site-packages/salt/utils/vt.py
