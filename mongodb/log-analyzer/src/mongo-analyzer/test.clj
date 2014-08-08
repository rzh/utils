;; gorilla-repl.fileformat = 1

;; **
;;; # Quick-Check/Generative Test/Property Based Testing
;;; 
;;; 
;; **

;; @@
(ns log-analyzer.test
  (:require [clojure.test.check :as tc]
            [clojure.test.check.generators :as gen]
            [monger.core :as mg]
            [monger.collection :as mc]
            [clojure.pprint :as p]
            [monger.operators :refer :all]
            [clojure.test.check.properties :as prop]))


(def db (mg/get-db (mg/connect) "test"))

;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;log-analyzer.test/db</span>","value":"#'log-analyzer.test/db"}
;; <=

;; @@
(defn add1
  [a]
  (if (= 100 a)
    99
    (+ a 1)))
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;log-analyzer.test/add1</span>","value":"#'log-analyzer.test/add1"}
;; <=

;; @@
(def add1-prop
  (prop/for-all [v gen/int]
                (= (+ 1 v) (add1 v))))

(p/pprint 
 (tc/quick-check 1000 add1-prop))

;; @@
;; ->
;;; {:result false,
;;;  :seed 1404323075608,
;;;  :failing-size 197,
;;;  :num-tests 798,
;;;  :fail [100],
;;;  :shrunk
;;;  {:total-nodes-visited 7, :depth 0, :result false, :smallest [100]}}
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; @@
;; test query shapes
;; :a int
(defn mongo-query-key 
  [_val _key db coll]

  (mc/remove db coll) ;; remove all doc

  (mc/ensure-index db coll (array-map _key 1) { :name "by-key" })

  (mc/insert db coll {_key _val} )
  
  (_key (mc/find-one-as-map db coll {_key _val})) ;; insert and return the inserted doc, take :key
  )


(defn mongo-query-map 
  [_val _key db coll]
  
  ;;- remove all doc in the collection
  (mc/remove db coll) 
  
  ;;- insert the doc
  (mc/insert db coll _val )
  
  ;;- return query results without _id
  (apply dissoc (mc/find-one-as-map db coll _val) [:_id])
  )


(defn mongo-test-vector 
  [_val db coll]
  
  ;;- remove all doc in the collection
  (mc/remove db coll) 
  
  ;;- insert the doc
  (mc/insert db coll {:a _val} )
  
  ;;- return query results without _id
  (:a (mc/find-one-as-map db coll {:a _val}))
  )


;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;log-analyzer.test/mongo-test-vector</span>","value":"#'log-analyzer.test/mongo-test-vector"}
;; <=

;; @@
(def mongo-insert-query-prop-key
  (prop/for-all [v gen/int]
                (= v (mongo-query-key v :a db "test1"))))

;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;log-analyzer.test/mongo-insert-query-prop-key</span>","value":"#'log-analyzer.test/mongo-insert-query-prop-key"}
;; <=

;; @@


(def mongo-insert-query-prop-map
  (prop/for-all [v (gen/map gen/keyword  gen/int)]
                (= v (mongo-query-map v :a db "test1"))))


(def mongo-insert-query-prop-vector 
  (prop/for-all [v (gen/not-empty (gen/vector gen/int))]
                	(= v (mongo-test-vector v db "test1"))
                ))
  
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;log-analyzer.test/mongo-insert-query-prop-vector</span>","value":"#'log-analyzer.test/mongo-insert-query-prop-vector"}
;; <=

;; **
;;; ## let's run some test
;; **

;; @@
(prn (tc/quick-check 100  mongo-insert-query-prop-key ))

;; @@
;; ->
;;; {:result true, :num-tests 100, :seed 1404323131943}
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; @@
(prn (tc/quick-check 100  mongo-insert-query-prop-map ))

;; @@
;; ->
;;; {:result true, :num-tests 100, :seed 1404323133985}
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; @@
(prn (tc/quick-check 100  mongo-insert-query-prop-vector ))
;; @@
;; ->
;;; {:result true, :num-tests 100, :seed 1404323137370}
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; **
;;; ### Test update with $pop
;; **

;; @@
(def db (mg/get-db (mg/connect) "test"))


(defn mongo-test-update-vector 
  [_val db coll]
  
  ;;- remove all doc in the collection
  (mc/remove db coll) 
  
  ;;- insert the doc
  (mc/insert db coll {:a _val} )
  
  ;;- use $push to add one to the array
  (mc/update db coll {:a _val} {$pop {:a -1}})

  ;;- return query results without _id
  (= (subvec _val 1) (:a (mc/find-one-as-map db coll {}))))



(def mongo-update-prop-vector 
  (prop/for-all [v (gen/not-empty (gen/vector gen/int))]
                	(mongo-test-update-vector v db "test1")
                ))

;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;log-analyzer.test/mongo-update-prop-vector</span>","value":"#'log-analyzer.test/mongo-update-prop-vector"}
;; <=

;; @@
(prn (tc/quick-check 50  mongo-update-prop-vector ))
;; @@
;; ->
;;; {:result true, :num-tests 50, :seed 1404323161785}
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; **
;;; ## GEO
;; **

;; @@
;;;; GEO
; http://grokbase.com/t/gg/clojure/143aqaajwa/ann-1st-public-release-of-thi-ng-geometry-toolkit-clj-cljs
; https://github.com/nakkaya/vector-2d
; https://groups.google.com/forum/#!topic/numerical-clojure/MTBxiN0GCTM
; https://astanin.github.io/clojure-math/clojure.math.geometry.html
; https://github.com/TheClimateCorporation/astro-algo
; cookbook chapeter 1.18 -> file:///Users/rui/git/clojure-cookbook/01_primitive-data/1-18_trigonometry.html
; http://crossclj.info/ns/cc.qbits/sextant/0.1.0/qbits.sextant.html#_PGeolocation


;; @@

;; @@

;; @@
