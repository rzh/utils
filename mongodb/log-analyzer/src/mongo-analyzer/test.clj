
(ns log-analyzer.test
  (:require [clojure.test.check :as tc]
            [clojure.test.check.generators :as gen]
            [monger.core :as mg]
            [monger.collection :as mc]
            [clojure.pprint :as p]
            [clojure.test.check.properties :as prop]))



(defn add1
  [a b]
  ; (if (= b 100) (prn "test"))
   (if (= 100 a)
     99
     (+ a 1)))

(def add1-prop
  (prop/for-all [v gen/int]
                (= (+ 1 v) (add1 v 100))))

(prn (tc/quick-check 1000 add1-prop))

(prn (add1 100 100))


;; test query shapes
;; :a int
(defn mongo-query-key [_val _key db coll]
  (mc/remove db coll) ;; remove all doc

  (mc/ensure-index db coll (array-map _key 1) { :name "by-key" })

  (mc/insert db coll {_key _val} )
  (_key (mc/find-one-as-map db coll {_key _val})) ;; insert and return the inserted doc, take :key
  )


(defn mongo-query-map [_val _key db coll]
  (mc/remove db coll) ;; remove all doc
  (mc/insert db coll _val )
  (apply dissoc (mc/find-one-as-map db coll _val) [:_id]) ;; insert and return the inserted doc, take :key
  )

(def db (mg/get-db (mg/connect) "test"))


(def mongo-insert-query-prop-map
  (prop/for-all [v (gen/map gen/keyword  gen/int)]
                (= v (mongo-query-map v :a db "test1"))))

(def mongo-insert-query-prop-key
  (prop/for-all [v (gen/one-of [(gen/vector gen/int)
;                                (gen/nat)
                                (gen/map gen/keyword gen/int)]) ]

                (= v (mongo-query-key v :a db "test1"))))


(prn (tc/quick-check 100000  mongo-insert-query-prop-map ))


(p/pprint (tc/quick-check 100  mongo-insert-query-prop-map ))
(def t (tc/quick-check 5000  mongo-insert-query-prop-key ))

(p/print-table t)


;;;; GEO
; http://grokbase.com/t/gg/clojure/143aqaajwa/ann-1st-public-release-of-thi-ng-geometry-toolkit-clj-cljs
; https://github.com/nakkaya/vector-2d
; https://groups.google.com/forum/#!topic/numerical-clojure/MTBxiN0GCTM
; https://astanin.github.io/clojure-math/clojure.math.geometry.html
; https://github.com/TheClimateCorporation/astro-algo
; cookbook chapeter 1.18 -> file:///Users/rui/git/clojure-cookbook/01_primitive-data/1-18_trigonometry.html
; http://crossclj.info/ns/cc.qbits/sextant/0.1.0/qbits.sextant.html#_PGeolocation

