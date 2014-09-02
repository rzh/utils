
# mongocmp

To compare output of mongo-perf and output to Go's benchcmp format for visualizeion

# How to use

First install mongocmp and my fork of benchviz

<pre>
go install github.com/rzh/utils/go/mongocmp
go install github.com/rzh/svgo/benchviz
</pre>

To compare test results
<pre>
go run mongocmp.go baseline.log new.log   | benchviz -vp=735 -h 1680 -w 1058 -line -vw 180 -title="sharded vs standalone" > sharded_perf.svg
</pre>

# How to read output

Here is the output SVG and how to read it
![Output](resources/mongo-perf-svg-instruction.png)

# Use -wiki output

mongocmp can directly output Jira/Wiki friendly output. Specify it with -wiki=true, such as

<pre>
go run mongocmp.go -wiki=true baseline.log new.log
</pre>

Output will be as follow with proper color code for failure/warning

<pre>
||benchmark                                         ||baseline ns/op     ||cv baseline                    ||new ns/op     ||cv new                         ||delta|
||Commands.CountsFullCollection_TH-5                |37626.40            |{color:orange}9.49%{color}      |38850.65       |1.41%                           |+3.25%|
||Commands.CountsIntIDRange_TH-5                    |21811.92            |2.36%                           |23640.21       |{color:orange}4.03%{color}      |+8.38%|
||Commands.DistinctWithIndexAndQuery_TH-5           |29360.73            |{color:orange}4.81%{color}      |31281.90       |{color:orange}4.24%{color}      |+6.54%|
||Commands.DistinctWithIndex_TH-5                   |30800.11            |{color:orange}4.03%{color}      |29892.79       |{color:orange}5.71%{color}      |-2.95%|
||Commands.DistinctWithoutIndexAndQuery_TH-5        |3003.73             |0.41%                           |3012.72        |0.44%                           |+0.30%|
||Commands.DistinctWithoutIndex_TH-5                |3198.69             |0.77%                           |3201.47        |0.68%                           |+0.09%|
||Commands.FindAndModifyInserts_TH-5                |20656.21            |{color:orange}9.07%{color}      |30879.80       |1.37%                           |+49.49%|
||Insert.EmptyCapped_TH-5                           |32511.31            |1.17%                           |25458.66       |0.26%                           |{color:red}-21.69%{col
or}|
||Insert.Empty_TH-5                                 |29678.01            |{color:orange}3.27%{color}      |30520.61       |1.28%                           |+2.84%|
</pre>

copy and paster and you are ready!
