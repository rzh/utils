
var functions = {}

var run_multiple_db = function () {
    return __run_multiple_db(false);
}

var run_multiple_db_index_field = function () {
    return __run_multiple_db(true);
}

var __run_multiple_db = function (indexField) {
  var ops = []; 
  
  for ( var i = 1; i <= 20; i++) {
    // drop collections
    var d = db.getSiblingDB("test"+i);
    d.getCollection("foo").drop();

    if ( indexField ) {
        d.getCollection("foo").ensureIndex({a: 1});
    }

    ops.push({ ns : "test"+i+".foo",
            op : "insert" ,
            doc : { a : { "#RAND_INT" : [ 0 , 100000000 ] } } ,
            writeCmd : true});
  }

  res = benchRun( {
      ops: ops,
      seconds : 20,
      totals : true,
      writeCmd : true,
      parallel : 20
      });

  return res;
}

var run_multiple_col = function () {
    return __run_multiple_col(false);
}
var run_multiple_col_index_field = function () {
    return __run_multiple_col(true);
}

var __run_multiple_col = function (indexField) {
  var ops = []; 
  var d = db.getSiblingDB("test_multi_col");
  
  for ( var i = 1; i <= 20; i++) {
    // drop collections
    d.getCollection("foo"+i).drop();

    if ( indexField ) {
        d.getCollection("foo"+i).ensureIndex({a: 1});
    }

    ops.push({ ns : "test_multi_col.foo"+i,
            op : "insert" ,
            doc : { a : { "#RAND_INT" : [ 0 , 100000000 ] } } ,
            writeCmd : true});
  }

  res = benchRun( {
      ops: ops,
      seconds : 20,
      totals : true,
      writeCmd : true,
      parallel : 20 });

  return res;
}

var run_single_col = function () {
    return __run_single_col(false);
}
var run_single_col_index_field = function () {
    return __run_single_col(true);
}

var __run_single_col = function (indexField) {
  // drop collections
  var d = db.getSiblingDB("test_single_col");
  d.getCollection("foo").drop();

  if ( indexField ) {
    d.getCollection("foo").ensureIndex({a: 1});
  }

  res = benchRun( {
      ops : [{
                ns : "test_single_col.foo",
                op : "insert" ,
                doc : { a : { "#RAND_INT" : [ 0 , 100000000 ] } } ,
                writeCmd : true }],
      seconds : 20,
      totals : true,
      writeCmd : true,
      parallel : 20
    });

  return res;
};

var run_tests = function () {
    var r;

    // print some server information
    print("\n");
    if (typeof  db.serverStatus().storageEngine != 'undefined') {
        print("storageEngine: " + db.serverStatus().storageEngine.name);
    }
    print("serverVersion: " + db.serverBuildInfo().version);

    //run multi-db
    for ( i = 0; i < 2; i++ ) {
        r = run_multiple_db();    
    }

    print("multi_db   : " + r.insert );
    
    //run multi-col
    for ( i = 0; i < 2; i++ ) {
        r = run_multiple_col();    
    }

    print("multi_col  : " + r.insert );
    
    //run single-col
    for ( i = 0; i < 2; i++ ) {
        r = run_single_col();    
    }

    print("single_col : " + r.insert );
}


var run_tests_index_field = function () {
    var r;

    // print some server information
    print("\nInsert with field indexed\n");
    if (typeof  db.serverStatus().storageEngine != 'undefined') {
        print("storageEngine: " + db.serverStatus().storageEngine.name);
    }
    print("serverVersion: " + db.serverBuildInfo().version);

    //run multi-db
    for ( i = 0; i < 2; i++ ) {
        r = run_multiple_db_index_field();    
    }

    print("multi_db   : " + r.insert );
    
    //run multi-col
    for ( i = 0; i < 2; i++ ) {
        r = run_multiple_col_index_field();    
    }

    print("multi_col  : " + r.insert );
    
    //run single-col
    for ( i = 0; i < 2; i++ ) {
        r = run_single_col_index_field();    
    }

    print("single_col : " + r.insert );
}

