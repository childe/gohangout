go build -o build/gohangout-test || exit 255
gohangout='build/gohangout-test'

em_print() {
  echo "\n======="
  echo $1
  echo "=======\n"
}

em_print 'test multi inputs&outputs'
tmpfile=$(mktemp)

$gohangout --config test/itest-1.yml > $tmpfile

wcl=`wc -l $tmpfile | awk '{print $1}'`
echo  "$wcl lines in output"

if [ "$wcl" != "4000" ]
then
	em_print  'should output 4000 docs!'
    exit 255
fi

em_print 'test simple filters with if condition'
$gohangout --config test/itest-2.yml > $tmpfile

wcl=`wc -l $tmpfile | awk '{print $1}'`
echo  "$wcl lines in output"

if [ "$wcl" != "2000" ]
then
	em_print  'should output 2000 docs!'
    exit 255
fi

em_print 'test output with if condition'
$gohangout --config test/itest-3.yml > $tmpfile

wcl=`wc -l $tmpfile | awk '{print $1}'`
echo  "$wcl lines in output"

if [ "$wcl" != "3000" ]
then
	em_print  'should output 3000 docs!'
    exit 255
fi

em_print 'test filtersFilter'
$gohangout --config test/itest-4.yml > $tmpfile

wcl=`wc -l $tmpfile | awk '{print $1}'`
echo  "$wcl lines in output"

if [ "$wcl" != "3000" ]
then
	em_print  'should output 3000 docs!'
    exit 255
fi

wcl=`grep tag1 $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl tag1 lines in output"

if [ "$wcl" != "3000" ]
then
	em_print  'tag1 should output 3000 docs!'
    exit 255
fi

wcl=`grep tag2 $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl tag2 lines in output"

if [ "$wcl" != "3000" ]
then
	em_print  'tag2 should output 3000 docs!'
    exit 255
fi

em_print 'test LinkMetricInFilters'

($gohangout --config test/itest-6.yml && sleep 1) | $gohangout --config test/LinkMetricInFilters.yml > $tmpfile

wcl=`grep count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "4" ]
then
	em_print  'metric should output 8 docs!'
    exit 255
fi

em_print 'test LinkMetric Filter 1: seperate'

($gohangout --config test/itest-6.yml && sleep 1) | $gohangout --config test/itest-6-2.yml > $tmpfile

wcl=`grep count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "4" ]
then
	em_print  'metric should output 8 docs!'
    exit 255
fi

($gohangout --config test/itest-6.yml && sleep 2) | $gohangout --config test/itest-6-2.yml > $tmpfile

wcl=`grep count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "4" ]
then
	em_print  'metric should output 3 docs!'
    exit 255
fi

wcl=`grep -v count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl raw lines in output"

if [ "$wcl" != "1000" ]
then
	em_print  'raw should output 1000 docs!'
    exit 255
fi

em_print 'test LinkMetric Filter 2: cumulative'

($gohangout --config test/itest-6.yml && sleep 1) | $gohangout --config test/itest-6-3.yml > $tmpfile

wcl=`grep count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "4" ]
then
	em_print  'metric should output 3 docs!'
    exit 255
fi

($gohangout --config test/itest-6.yml && sleep 2) | $gohangout --config test/itest-6-3.yml > $tmpfile

wcl=`grep count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "8" ]
then
	em_print  'metric should output 8 docs!'
    exit 255
fi

wcl=`grep -v count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl raw lines in output"

if [ "$wcl" != "1000" ]
then
	em_print  'raw should output 1000 docs!'
    exit 255
fi

em_print 'test LinkMetric Filter 3: seperate'

($gohangout --config test/itest-6.yml && sleep 2 && $gohangout --config test/itest-6.yml && sleep 2) | $gohangout --config test/itest-6-2.yml > $tmpfile

wcl=`grep count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "8" ]
then
	em_print  'metric should output 8 docs!'
    exit 255
fi

wcl=`grep -v count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl raw lines in output"

if [ "$wcl" != "2000" ]
then
	em_print  'raw should output 2000 docs!'
    exit 255
fi

em_print 'test LinkMetric Filter 4: cumulative'

($gohangout --config test/itest-6.yml && sleep 1 && $gohangout --config test/itest-6.yml && sleep 2) | $gohangout --config test/itest-6-4.yml > $tmpfile

wcl=`grep count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "20" ]
then
	em_print  'metric should output 20 docs!'
    exit 255
fi

wcl=`grep -v count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl raw lines in output"

if [ "$wcl" != "2000" ]
then
	em_print  'raw should output 2000 docs!'
    exit 255
fi

em_print 'test LinkStatsMetrcik Filter'

($gohangout --config test/itest-6.yml && sleep 1 && $gohangout --config test/itest-6.yml && sleep 2) | $gohangout --config test/itest-7-1.yml > $tmpfile

wcl=`grep count $tmpfile | wc -l | awk '{print $1}'`
echo  "$wcl metric lines in output"

if [ "$wcl" != "20" ]
then
	em_print  'metric should output 20 docs!'
    exit 255
fi

# test tcp input/output
test/itest-tcp.sh
if [ "$?" != "0" ]
then
    exit $?
fi

em_print 'pass :)'
