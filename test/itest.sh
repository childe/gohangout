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
