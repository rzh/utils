#!/bin/bash 

go run main.go no-shard.log shard-id-hash.log   | benchviz -vp=710 -h 1680 -w 1058 -line -vw 180 -title="_id:hashed vs standalone" > id-hash.svg
go run main.go no-shard.log shard-id-1.log      | benchviz -vp=710 -h 1680 -w 1058 -line -vw 180 -title="_id:1 vs standalone" > id-1.svg
go run main.go shard-id-1.log shard-id-hash.log | benchviz -vp=710 -h 1680 -w 1058 -line -vw 180 -title="_id:hashed vs _id:1" > hash-vs-id1.svg
