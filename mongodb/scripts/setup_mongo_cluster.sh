#/bin/bash

initMongo()
{
  # Setup MongoDB data & logs dirs
  mkdir -p $1/{data,logs}/{standalone,rs-{0,1,2},sharded/{shard{a,b}-0,config-{0,1,2},mongos}}
}

#startMongo(rootDir,type,port,options)
startMongo()
{
    local rootDir=$1
    local ttype=$2
    local port=$3
    local options=$4

    local mongoCmd="$rootDir/bin/mongod \
    --port $port \
    --fork \
    --dbpath $rootDir/data/$ttype/ \
    --logpath $rootDir/logs/$ttype/mongod.log \
    --smallfiles \
    --nohttpinterface \
    $options"

    echo "Starting: $mongoCmd"
    $mongoCmd
}

#stopandcleanMongo(rootDir)
stopandcleanMongo()
{
    echo "Killing mongod"
    sudo killall -9 mongod
    echo "Killing mongos"
    sudo killall -9 mongos
    echo "Removing $1/{data,logs}/"
    sudo rm -fr $1/{data,logs}/
}


# start relicaset(rootDir)
startReplicaSet()
{
    mongoRoot=$1
    hostname=`hostname`
    # hostname="192.168.19.101"
    startMongo $mongoRoot rs-0 27017  "--replSet rs --oplogSize=50"
    startMongo $mongoRoot rs-1 27018  "--replSet rs --oplogSize=50"
    startMongo $mongoRoot rs-2 27019  "--replSet rs --oplogSize=50"
    # sleep 10
    $mongoRoot/bin/mongo --port 27017 \
      --eval "rs.initiate();\
      sleep(2000);\
      rs.add(\"${hostname}:27018\");\
      rs.add(\"${hostname}:27019\");\
      rs.slaveOk();\
      cfg = rs.conf();\
      printjson(cfg);\
      cfg.members[0].priority = 1;\
      cfg.members[1].priority = 0.5;\
      cfg.members[2].priority = 0.5;\
      rs.reconfig(cfg);\
      sleep(5000);\
      printjson(rs.status())"
      # sleep 10
}

createGSSAPI_users() {
    mongoRoot=$1

    $mongoRoot/bin/mongo --port 27017 --verbose <<! 
use \$external;
db.createUser( { user: "gssapitest@MONGOTEST.COM", roles: [ {role: "readWrite", db: "test" }, {role: "read", db: "admin" }, {role: "clusterAdmin", db: "admin" } ] } );
db.createUser( { user: "gssapitest1@MONGOTEST.COM", roles: [ {role: "readWrite", db: "test" }] } );
sleep(5000);
printjson(db.getUsers());
!
}


restartMongods_with_gssapi() {
    mongoRoot=$1

    echo "Killing mongod"
    sudo killall -9 mongod

    echo "start mongod in GSSAPI"
    startMongo $mongoRoot rs-0 27017  "--auth --setParameter=authenticationMechanisms=GSSAPI --replSet rs --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile"
    startMongo $mongoRoot rs-1 27018  "--auth --setParameter=authenticationMechanisms=GSSAPI --replSet rs --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile"
    startMongo $mongoRoot rs-2 27019  "--auth --setParameter=authenticationMechanisms=GSSAPI --replSet rs --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile"
}

startShardedCluster() {
    mongoRoot=$1

    # first start config server
    startMongo $mongoRoot "sharded/config-0" 27100  "--configsvr --oplogSize=50"
    startMongo $mongoRoot "sharded/config-1" 27101  "--configsvr --oplogSize=50"
    startMongo $mongoRoot "sharded/config-2" 27102  "--configsvr --oplogSize=50"
    sleep 2

    # now start mongos
    $mongoRoot/bin/mongos --fork --configdb localhost:27100,localhost:27101,localhost:27102 --logpath="/home/vagrant/mongodb/logs/sharded/mongos/mongod.log"

    # now start mongod
    startMongo $mongoRoot "sharded/sharda-0" 27217  "--oplogSize=50"
    startMongo $mongoRoot "sharded/shardb-0" 27218  "--oplogSize=50"

    # now configure shard
    $mongoRoot/bin/mongo --port 27017 --verbose <<!
sh.addShard("localhost:27217");
sh.addShard("localhost:27218");
sh.enableSharding("test");
use \$external;
db.createUser( { user: "gssapitest@MONGOTEST.COM", roles: [ {role: "readWrite", db: "test" }, {role: "read", db: "admin" }, {role: "clusterAdmin", db: "admin" } ] } );
db.createUser( { user: "gssapitest1@MONGOTEST.COM", roles: [ {role: "readWrite", db: "test" }] } );
sleep(5000);
printjson(db.getUsers());
!

    $mongoRoot/bin/mongo --port 27217 --verbose <<!
use \$external;
db.createUser( { user: "gssapitest@MONGOTEST.COM", roles: [ {role: "readWrite", db: "test" }, {role: "read", db: "admin" }, {role: "clusterAdmin", db: "admin" } ] } );
db.createUser( { user: "gssapitest1@MONGOTEST.COM", roles: [ {role: "readWrite", db: "test" }] } );
sleep(5000);
printjson(db.getUsers());
!

    $mongoRoot/bin/mongo --port 27218 --verbose <<!
use \$external;
db.createUser( { user: "gssapitest@MONGOTEST.COM", roles: [ {role: "readWrite", db: "test" }, {role: "read", db: "admin" }, {role: "clusterAdmin", db: "admin" } ] } );
db.createUser( { user: "gssapitest1@MONGOTEST.COM", roles: [ {role: "readWrite", db: "test" }] } );
sleep(5000);
printjson(db.getUsers());
!
}

restartShardedCluster_with_gssapi() {
    mongoRoot=$1

    # first start config server
    startMongo $mongoRoot "sharded/config-0" 27100  "--configsvr --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile --auth --setParameter=authenticationMechanisms=GSSAPI"
    startMongo $mongoRoot "sharded/config-1" 27101  "--configsvr --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile --auth --setParameter=authenticationMechanisms=GSSAPI"
    startMongo $mongoRoot "sharded/config-2" 27102  "--configsvr --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile --auth --setParameter=authenticationMechanisms=GSSAPI"
    sleep 2

    # now start mongos
    $mongoRoot/bin/mongos --fork --configdb localhost:27100,localhost:27101,localhost:27102 \
      --logpath="/home/vagrant/mongodb/logs/sharded/mongos/mongod.log" \
      --setParameter authenticationMechanisms=GSSAPI \
      --keyFile=$mongoRoot/rs.keyfile 

    # now start mongod
    startMongo $mongoRoot "sharded/sharda-0" 27217  "--oplogSize=50 --auth --keyFile=$mongoRoot/rs.keyfile --setParameter=authenticationMechanisms=GSSAPI"
    startMongo $mongoRoot "sharded/shardb-0" 27218  "--oplogSize=50 --auth --keyFile=$mongoRoot/rs.keyfile --setParameter=authenticationMechanisms=GSSAPI"

}

restartShardedCluster() {
    mongoRoot=$1

    # first start config server
    startMongo $mongoRoot "sharded/config-0" 27100  "--configsvr --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile"
    startMongo $mongoRoot "sharded/config-1" 27101  "--configsvr --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile"
    startMongo $mongoRoot "sharded/config-2" 27102  "--configsvr --oplogSize=50 --keyFile=$mongoRoot/rs.keyfile"
    sleep 2

    # now start mongos
    $mongoRoot/bin/mongos --fork --configdb localhost:27100,localhost:27101,localhost:27102 \
      --logpath="/home/vagrant/mongodb/logs/sharded/mongos/mongod.log" \
      --keyFile=$mongoRoot/rs.keyfile 

    # now start mongod
    startMongo $mongoRoot "sharded/sharda-0" 27217  "--oplogSize=50 --auth --keyFile=$mongoRoot/rs.keyfile"
    startMongo $mongoRoot "sharded/shardb-0" 27218  "--oplogSize=50 --auth --keyFile=$mongoRoot/rs.keyfile"

}


cd $HOME
mongoRoot=`pwd`/mongodb

if [ -f $mongoRoot/rs.keyfile ]; then
  rm -f $mongoRoot/rs.keyfile
fi

openssl rand -base64 741 > $mongoRoot/rs.keyfile
chmod 400 $mongoRoot/rs.keyfile

stopandcleanMongo $mongoRoot
initMongo $mongoRoot

case $1 in
  replicaset)
    startReplicaSet $mongoRoot
    createGSSAPI_users $mongoRoot
    restartMongods_with_gssapi $mongoRoot
    ;;

  sharded)
    echo "sharded!!!!"
    startShardedCluster $mongoRoot
    sleep 5
    echo ""
    echo ""
    echo ""

    echo "Killing mongod"
    sudo killall -9 mongod
    echo "Killing mongos"
    sudo killall -9 mongos

    restartShardedCluster_with_gssapi $mongoRoot
    ;;

esac


