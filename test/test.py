# -*- coding: utf-8 -*-

import os
import subprocess
import json

msg = [
    # grok
    '2006-01-02T15:04:05 liujia 200',
    '2006-01-02T15:04:05 childe 200',
    # drop
    '2006-01-02T15:04:05 null 200',
    # grok match 2
    '2006-01-02T15:04:05 200 debug',
    # add
    json.dumps({'@timestamp': '2006-01-02T15:04:05', 'client': 'liujia', 'stored': {'message': 'hello gohangout'}})
       ]

p1 = subprocess.Popen(["build/gohangout", "--config", "test/t.yml"],
                      # p1 = subprocess.Popen(["cat"],
                      stdin=subprocess.PIPE, stdout=subprocess.PIPE)
for m in msg:
    p1.stdin.write(m)
    p1.stdin.write('\n')

output = p1.communicate()[0]
print output

expectation = '''{"@timestamp":"2006-01-02T07:04:05Z","a":{"b":null},"b":"xyz","name":"liujia","status":"200","xxx":"xxx","yyy":null,"zzz":null}
{"@timestamp":"2006-01-02T07:04:05Z","a":"xyz","name":"childe","status":"200","xxx":"xxx","yyy":null,"zzz":null}
{"@timestamp":"2006-01-02T07:04:05Z","a":{"b":null},"loglevel":"debug","status":"200","xxx":"xxx","yyy":null,"zzz":null}
{"@timestamp":"2006-01-02T15:04:05","a":{"b":"hello gohangout"},"client":"liujia","stored":{"message":"hello gohangout"},"xxx":"xxx","yyy":"liujia","zzz":"hello gohangout"}
'''


for a, b in zip(output.strip().split('\n'), expectation.strip().split('\n')):
    print '=' * 40
    print a
    print b
    print json.loads(a) == json.loads(b)
