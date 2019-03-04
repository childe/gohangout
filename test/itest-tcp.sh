go build -o build/gohangout-test

em_print() {
  echo "\n======="
  echo $1
  echo "=======\n"
}

em_print 'test tcp input/output'

tmpfile=$(mktemp)

nohup build/gohangout-test --config test/itest-tcpinput.yml > $tmpfile &
sleep 1

build/gohangout-test --config test/itest-tcpoutput.yml
sleep 2

ps -ef | grep 'build/gohangout-test --config test/itest-tcpinput.yml' | grep -v grep | awk '{print $2}' | xargs kill

wcl=$(wc -l $tmpfile | awk '{print $1}')


if [ "$wcl" != "200000" ]
then
	em_print  'tcp input/output should create 200000 docs!'
    exit 255
else
    em_print 'tcp plugin passes'
fi

rm $tmpfile

