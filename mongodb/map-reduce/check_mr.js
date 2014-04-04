

mr_db = db.getSiblingDB("mr_results")

for(i =0; i < 200000; i++) {
	// print("check " + i)
	single_result = mr_db.totals.findOne({_id: i})
	concurrent_result = mr_db.totals_concurrent.findOne({_id: i})

	// when parallelly run 8 MR jobs on the same dataset, expecting 8 time in results
	if( concurrent_result.value * 1 != single_result.value * 8 ) {
		print("mismatch for doc " + i + " single got  " + single_result.value + " and concurrent get  " + concurrent_result.value)
	}
}
