
// single threaded MR job
mr_db = db.getSiblingDB("mr")

fmap    = function() { emit( this.uid, this.amount ); },
freduce = function(k, v) { return Array.sum(v) },
query   = { out: { reduce: "totals", db: "mr_results", nonAtomic: true} }

re = mr_db.mr.mapReduce( fmap, freduce, query)

printjson(re)
