go build -o build/gohangout-test

em_print() {
  echo "\n======="
  echo $1
  echo "=======\n"
}

em_print 'test multi inputs&outputs'
build/gohangout-test --config test/itest-1.yml > test/gohangout-test.output.txt

wcl=`wc -l test/gohangout-test.output.txt | awk '{print $1}'`
echo  "$wcl lines in output"

if [ "$wcl" != "4000" ]
then
	em_print  'should output 4000 docs!'
    exit 255
fi

em_print 'test simple filters with if condition'
build/gohangout-test --config test/itest-2.yml > test/gohangout-test.output.txt

wcl=`wc -l test/gohangout-test.output.txt | awk '{print $1}'`
echo  "$wcl lines in output"

if [ "$wcl" != "2000" ]
then
	em_print  'should output 2000 docs!'
    exit 255
fi

em_print 'test output with if condition'
build/gohangout-test --config test/itest-3.yml > test/gohangout-test.output.txt

wcl=`wc -l test/gohangout-test.output.txt | awk '{print $1}'`
echo  "$wcl lines in output"

if [ "$wcl" != "3000" ]
then
	em_print  'should output 3000 docs!'
    exit 255
fi

em_print 'test filterFilter'
build/gohangout-test --config test/itest-4.yml > test/gohangout-test.output.txt

wcl=`wc -l test/gohangout-test.output.txt | awk '{print $1}'`
echo  "$wcl lines in output"

if [ "$wcl" != "3000" ]
then
	em_print  'should output 3000 docs!'
    exit 255
fi

wcl=`grep tag1 test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl tag1 lines in output"

if [ "$wcl" != "3000" ]
then
	em_print  'tag1 should output 3000 docs!'
    exit 255
fi

wcl=`grep tag2 test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl tag2 lines in output"

if [ "$wcl" != "3000" ]
then
	em_print  'tag2 should output 3000 docs!'
    exit 255
fi

em_print 'test LinkMetrcik Filter 1: seperate'

(build/gohangout-test --config test/itest-6.yml && sleep 1) | build/gohangout-test --config test/itest-6-2.yml > test/gohangout-test.output.txt

wcl=`grep count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "4" ]
then
	em_print  'metric should output 8 docs!'
    exit 255
fi

(build/gohangout-test --config test/itest-6.yml && sleep 2) | build/gohangout-test --config test/itest-6-2.yml > test/gohangout-test.output.txt

wcl=`grep count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "4" ]
then
	em_print  'metric should output 3 docs!'
    exit 255
fi

wcl=`grep -v count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl raw lines in output"

if [ "$wcl" != "1000" ]
then
	em_print  'raw should output 1000 docs!'
    exit 255
fi

em_print 'test LinkMetrcik Filter 2: cumulative'

(build/gohangout-test --config test/itest-6.yml && sleep 1) | build/gohangout-test --config test/itest-6-3.yml > test/gohangout-test.output.txt

wcl=`grep count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "4" ]
then
	em_print  'metric should output 3 docs!'
    exit 255
fi

(build/gohangout-test --config test/itest-6.yml && sleep 2) | build/gohangout-test --config test/itest-6-3.yml > test/gohangout-test.output.txt

wcl=`grep count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "8" ]
then
	em_print  'metric should output 8 docs!'
    exit 255
fi

wcl=`grep -v count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl raw lines in output"

if [ "$wcl" != "1000" ]
then
	em_print  'raw should output 1000 docs!'
    exit 255
fi

em_print 'test LinkMetrcik Filter 3: seperate'

(build/gohangout-test --config test/itest-6.yml && sleep 2 && build/gohangout-test --config test/itest-6.yml && sleep 2) | build/gohangout-test --config test/itest-6-2.yml > test/gohangout-test.output.txt

wcl=`grep count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "8" ]
then
	em_print  'metric should output 8 docs!'
    exit 255
fi

wcl=`grep -v count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl raw lines in output"

if [ "$wcl" != "2000" ]
then
	em_print  'raw should output 2000 docs!'
    exit 255
fi

em_print 'test LinkMetrcik Filter 4: cumulative'

(build/gohangout-test --config test/itest-6.yml && sleep 1 && build/gohangout-test --config test/itest-6.yml && sleep 2) | build/gohangout-test --config test/itest-6-4.yml > test/gohangout-test.output.txt

wcl=`grep count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "20" ]
then
	em_print  'metric should output 20 docs!'
    exit 255
fi

wcl=`grep -v count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl raw lines in output"

if [ "$wcl" != "2000" ]
then
	em_print  'raw should output 2000 docs!'
    exit 255
fi

em_print 'test LinkStatsMetrcik Filter'

(build/gohangout-test --config test/itest-6.yml && sleep 1 && build/gohangout-test --config test/itest-6.yml && sleep 2) | build/gohangout-test --config test/itest-7-1.yml > test/gohangout-test.output.txt

wcl=`grep count test/gohangout-test.output.txt | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "20" ]
then
	em_print  'metric should output 20 docs!'
    exit 255
fi

echo 'ok'
