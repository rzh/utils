#!/bin/bash 

go run mongocmp.go no-shard.log shard-id-hash.log   | benchviz -vp=735 -h 1680 -w 1058 -line -vw 180 -title="_id:hashed vs standalone" > id-hash.svg
go run mongocmp.go no-shard.log shard-id-1.log      | benchviz -vp=735 -h 1680 -w 1058 -line -vw 180 -title="_id:1 vs standalone" > id-1.svg
go run mongocmp.go shard-id-1.log shard-id-hash.log | benchviz -vp=735 -h 1680 -w 1058 -line -vw 180 -title="_id:hashed vs _id:1" > hash-vs-id1.svg
