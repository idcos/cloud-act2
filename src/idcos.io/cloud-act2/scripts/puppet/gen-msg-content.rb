# author: mix
# modified: damon
# history:
# 2018-09-03： change to caller to cloud-act2

require 'json'
require 'digest'
require 'securerandom'
require 'base64'


## read stdin for json

json = ARGF.read
data = JSON.parse(json)

cmd = data['cmd']
requestid = data['requestid']


user = 'root'
#cmd = $cmd
psk = 'a36cd839414370e10fd281b8a38a4f48'
#requestid = $requestid
timeout = 3602

caller = 'cloud-act2'

msg = {
    agent: 'shell',
    action: 'run',
    caller: caller,
    data: {
        type: 'cmd',
        user: user,
        command: cmd,
        process_result: true,
        environment: ''
    }
}

req = {
    body: Marshal.dump(msg),
    senderid: caller,
    requestid: requestid,
    filter: {
        "agent" => ['shell'],
        collective: 'mcollective'
    },
    collective: 'mcollective',
    agent: 'shell',
    callerid: "cert=#{caller}",
    ttl: 3600,
    msgtime: Time.now.utc.to_i,
    hash: Digest::MD5.hexdigest(Marshal.dump(msg).to_s + psk)
}
STDOUT.write Marshal.dump(req).bytes.to_a
