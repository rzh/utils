(ns mongo-analyzer.gen
  (:require [monger.core :as mg]
            [monger.collection :as mc]
            [clojure.pprint :as p]))

(use 'criterium.core)

;; some study on how to create date from query shape

(def q-shape {:a :int :b :string })

(prn q-shape)


;; ╰─> test
(defn gen-map-from-shape
  "to create a map from shape"
  [shape n]
  ; to file the all map vals with n
    (into {} (for [[k v] shape]
               (if (map? v)
                 [k (gen-map-from-shape v n)]
                 [k n]))))


(defn gen-docs
  "create raw data for query shape"
  [shape n start]
  (for [x (range start (+ start n))]
    (assoc (gen-map-from-shape shape x) :_id x)
    ))

(p/pprint (gen-docs {:a 100 :b {:b1 100 :b2 100}} 20 10))


;;;----  test ----

;; 0 - loop over a range

(defn run-test!
  "run a perf test for given n number of doc,
  return perf result.

  Input:
     shape - doc shape
     n     - number of docs
     m     - run test for m times
     db    - mongodb
     coll  - collection
  "
  [shape n m db coll]

  ;; insert n doc into collection

  (mc/remove db coll) ;; remove all doc

  (for [x (range n)]
    ;; insert doc
    (mc/insert db coll
               (gen-map-from-shape shape x) :_id x)
    )

  ;; run test for x times
  (bench (mc/find-one db coll {:_id (rand-int n)}))
  )


;;; END


(def db (mg/get-db (mg/connect) "test"))

(run-test! {:a 100} 1000 1 db "foo")

(bench (Thread/sleep 1000))
