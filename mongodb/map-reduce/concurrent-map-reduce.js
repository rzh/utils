
mr_db = db.getSiblingDB("mr")

var numThreads = 8

threads=[]

mr_func = function(i) {
    var mr_db = db.getSiblingDB("mr")
	var fmap    = function() { emit( this.uid, this.amount ); }
	var freduce = function(k, v) { return Array.sum(v) }
	var out   = { out: {reduce: "totals_concurrent", db: "mr_results", nonAtomic: true}}
    t = mr_db.mr.mapReduce( fmap, freduce, out)

	print("thread -- ", i)
    printjson(t)
}

for(j =0; j < numThreads; j++){
	var t = new ScopedThread(mr_func, j);
	threads.push(t);
	t.start()
}

for (j =0; j < numThreads; j++) { var t = threads[j]; t.join(); printjson(t.returnData()); }

