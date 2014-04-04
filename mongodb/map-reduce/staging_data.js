
mr_db = db.getSiblingDB("mr")
mr_db.dropDatabase()

mr_db.mr.ensureIndex({uid: 1})
var numThreads = 10  // parallel writing with 10 threads

threads=[]

insert10k = function() {
    var mr_db = db.getSiblingDB("mr")
	for(i =0; i < 200000; i++) {
		for(k = 0; k < 5; k ++) {
			// share change to bulk insert to further speed up
			mr_db.mr.insert({uid: i, amount: Math.floor(Math.random()*1000000), status: k})
		}
	}
}

for(j =0; j < numThreads; j++){
	var t = new ScopedThread(insert10k);
	threads.push(t);
	t.start()
}

for (j =0; j < numThreads; j++) { var t = threads[j]; t.join(); printjson(t.returnData()); }

