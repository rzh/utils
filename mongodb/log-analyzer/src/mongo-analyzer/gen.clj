;; gorilla-repl.fileformat = 1

;; **
;;; # GEN
;;; 
;;; Perf test from query log
;; **

;; @@
(ns mongo-analyzer.gen
  (:require [monger.core :as mg]
            [monger.collection :as mc]
            [monger.operators :refer :all]
            [clojure.pprint :as p]))

(use 'criterium.core)

;; some study on how to create date from query shape

(def q-shape {:a :int :b :string })

(prn q-shape)
;; @@
;; ->
;;; {:b :string, :a :int}
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; **
;;; ## a same query from log file
;; **

;; @@
(def query-1 {:q {:OP_query {:status "OPEN" :cid "ObjectId('538fa43177143b548c76ecfd')" :cre {:OP_lte "new-Date(1401923435593)"}} :OP_orderby {:cre "-1"}} 
              :query-plan "IXSCAN"})

(p/pprint query-1)
;; @@
;; ->
;;; {:q
;;;  {:OP_query
;;;   {:cid &quot;ObjectId(&#x27;538fa43177143b548c76ecfd&#x27;)&quot;,
;;;    :cre {:OP_lte &quot;new-Date(1401923435593)&quot;},
;;;    :status &quot;OPEN&quot;},
;;;   :OP_orderby {:cre &quot;-1&quot;}},
;;;  :query-plan &quot;IXSCAN&quot;}
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; @@
;; ╰─> test
(defn gen-map-from-shape
  "to create a map from shape"
  [shape n & {:keys [query]
              :or {query false}}]
  ; to file the all map vals with n
    (into {} (for [[k v] shape]
               (cond 
                (map? v)
                 [k (gen-map-from-shape v n)]
                (vector? v) [k (cond 
                             	  (some #(= :rand-int %) v)
                                	(if (and query (some #(= :lte %) v))
                                      {$lte (rand-int 1000)}
                                      (rand-int 1000))
				                     
                			      (some #(= :rand %) v)
				                     (rand 1000)
                                  :else 10001)
                             ]
                (= :ObjectId v) 
                [k (str "ObjectId(" n ")")]
                
                (= :string v)
                [k (str "string_" n)]
                :else 
                 [k n]))))


(defn gen-index-map-from-schema 
  [schema]
  (into {} (for [[k v] schema]
             (cond
              (map? v)  [k (gen-index-map-from-schema)]
              :else     [k 1]
              ))))
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/gen-index-map-from-schema</span>","value":"#'mongo-analyzer.gen/gen-index-map-from-schema"}
;; <=

;; @@
(defn gen-docs
  "create raw data for query shape"
  [shape n start & {:keys [query]
                    :or {query false}}]
  (for [x (range start (+ start n))]
    (assoc (gen-map-from-shape shape x :query query) :_id x)
    ))

(p/pprint (gen-docs {:a 100 :b {:b1 100 :b2 100}} 20 10))


;; @@
;; ->
;;; ({:_id 10, :b {:b2 10, :b1 10}, :a 10}
;;;  {:_id 11, :b {:b2 11, :b1 11}, :a 11}
;;;  {:_id 12, :b {:b2 12, :b1 12}, :a 12}
;;;  {:_id 13, :b {:b2 13, :b1 13}, :a 13}
;;;  {:_id 14, :b {:b2 14, :b1 14}, :a 14}
;;;  {:_id 15, :b {:b2 15, :b1 15}, :a 15}
;;;  {:_id 16, :b {:b2 16, :b1 16}, :a 16}
;;;  {:_id 17, :b {:b2 17, :b1 17}, :a 17}
;;;  {:_id 18, :b {:b2 18, :b1 18}, :a 18}
;;;  {:_id 19, :b {:b2 19, :b1 19}, :a 19}
;;;  {:_id 20, :b {:b2 20, :b1 20}, :a 20}
;;;  {:_id 21, :b {:b2 21, :b1 21}, :a 21}
;;;  {:_id 22, :b {:b2 22, :b1 22}, :a 22}
;;;  {:_id 23, :b {:b2 23, :b1 23}, :a 23}
;;;  {:_id 24, :b {:b2 24, :b1 24}, :a 24}
;;;  {:_id 25, :b {:b2 25, :b1 25}, :a 25}
;;;  {:_id 26, :b {:b2 26, :b1 26}, :a 26}
;;;  {:_id 27, :b {:b2 27, :b1 27}, :a 27}
;;;  {:_id 28, :b {:b2 28, :b1 28}, :a 28}
;;;  {:_id 29, :b {:b2 29, :b1 29}, :a 29})
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; **
;;; ### --- test ---
;; **

;; @@
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
  (mc/drop-indexes db coll)
  (mc/ensure-index db coll (gen-index-map-from-schema shape))


  (doseq [x (range 0 1000)]
    (mc/insert db coll
               (assoc (gen-map-from-shape shape x) :_id x)))

  ;; run test for x times
  (benchmark (mc/find-one db coll (gen-map-from-shape shape (rand-int n) :query true)) nil)
  )


;;; END
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/run-test!</span>","value":"#'mongo-analyzer.gen/run-test!"}
;; <=

;; **
;;; 
;; **

;; **
;;; ## Test	
;; **

;; @@
(def db (mg/get-db (mg/connect) "test"))
(def coll "test")
;; (bench (Thread/sleep 1000))

;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/coll</span>","value":"#'mongo-analyzer.gen/coll"}
;; <=

;; @@
(def db (mg/get-db (mg/connect) "test"))

(def re
  (into {} 
        (for [i [1 5000 10000 50000 100000 200000 400000 1000000]] 
          {(keyword (str i)) (run-test! {:a 10} i 1 db "foo")}))
  )
;; @@
;; ->
;;; WARNING: Final GC required 1.45134916352533 % of runtime
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/db</span>","value":"#'mongo-analyzer.gen/db"}
;; <=

;; @@
(defn to-us [u] (* 100000 u))
(defn take-mean [r] (get (:mean r) 0))

(def results (into {} (for [[k, v] re]
  [k (* 1000000 (get (:mean v) 0))]
  )))
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/results</span>","value":"#'mongo-analyzer.gen/results"}
;; <=

;; @@
results
;; @@
;; =>
;;; {"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:1</span>","value":":1"},{"type":"html","content":"<span class='clj-double'>58.82900222841766</span>","value":"58.82900222841766"}],"value":"[:1 58.82900222841766]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:5000</span>","value":":5000"},{"type":"html","content":"<span class='clj-double'>58.83526940687395</span>","value":"58.83526940687395"}],"value":"[:5000 58.83526940687395]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:10000</span>","value":":10000"},{"type":"html","content":"<span class='clj-double'>59.80187727193341</span>","value":"59.80187727193341"}],"value":"[:10000 59.80187727193341]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:50000</span>","value":":50000"},{"type":"html","content":"<span class='clj-double'>58.94269362233496</span>","value":"58.94269362233496"}],"value":"[:50000 58.94269362233496]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:100000</span>","value":":100000"},{"type":"html","content":"<span class='clj-double'>58.72081656540282</span>","value":"58.72081656540282"}],"value":"[:100000 58.72081656540282]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:200000</span>","value":":200000"},{"type":"html","content":"<span class='clj-double'>59.55537611612778</span>","value":"59.55537611612778"}],"value":"[:200000 59.55537611612778]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:400000</span>","value":":400000"},{"type":"html","content":"<span class='clj-double'>58.80980503615203</span>","value":"58.80980503615203"}],"value":"[:400000 58.80980503615203]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:1000000</span>","value":":1000000"},{"type":"html","content":"<span class='clj-double'>59.08874797551242</span>","value":"59.08874797551242"}],"value":"[:1000000 59.08874797551242]"}],"value":"{:1 58.82900222841766, :5000 58.83526940687395, :10000 59.80187727193341, :50000 58.94269362233496, :100000 58.72081656540282, :200000 59.55537611612778, :400000 58.80980503615203, :1000000 59.08874797551242}"}
;; <=

;; @@
(run-test! {:a 100} 5000000 1 db "foo")
;; @@
;; =>
;;; {"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:overhead</span>","value":":overhead"},{"type":"html","content":"<span class='clj-double'>1.662542005421643E-9</span>","value":"1.662542005421643E-9"}],"value":"[:overhead 1.662542005421643E-9]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:samples</span>","value":":samples"},{"type":"list-like","open":"<span class='clj-lazy-seq'>(<span>","close":"<span class='clj-lazy-seq'>)</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-long'>1008929870</span>","value":"1008929870"},{"type":"html","content":"<span class='clj-long'>1006691623</span>","value":"1006691623"},{"type":"html","content":"<span class='clj-long'>1007274197</span>","value":"1007274197"},{"type":"html","content":"<span class='clj-long'>1005775539</span>","value":"1005775539"},{"type":"html","content":"<span class='clj-long'>1009803490</span>","value":"1009803490"},{"type":"html","content":"<span class='clj-long'>1014054834</span>","value":"1014054834"},{"type":"html","content":"<span class='clj-long'>1008616404</span>","value":"1008616404"},{"type":"html","content":"<span class='clj-long'>1006501909</span>","value":"1006501909"},{"type":"html","content":"<span class='clj-long'>1014588731</span>","value":"1014588731"},{"type":"html","content":"<span class='clj-long'>1042267135</span>","value":"1042267135"},{"type":"html","content":"<span class='clj-long'>1005476415</span>","value":"1005476415"},{"type":"html","content":"<span class='clj-long'>1007543463</span>","value":"1007543463"},{"type":"html","content":"<span class='clj-long'>1009080563</span>","value":"1009080563"},{"type":"html","content":"<span class='clj-long'>1006871413</span>","value":"1006871413"},{"type":"html","content":"<span class='clj-long'>1006107920</span>","value":"1006107920"},{"type":"html","content":"<span class='clj-long'>1005764034</span>","value":"1005764034"},{"type":"html","content":"<span class='clj-long'>1007602449</span>","value":"1007602449"},{"type":"html","content":"<span class='clj-long'>1007358811</span>","value":"1007358811"},{"type":"html","content":"<span class='clj-long'>1010391196</span>","value":"1010391196"},{"type":"html","content":"<span class='clj-long'>1009402876</span>","value":"1009402876"},{"type":"html","content":"<span class='clj-long'>1010927691</span>","value":"1010927691"},{"type":"html","content":"<span class='clj-long'>1041252800</span>","value":"1041252800"},{"type":"html","content":"<span class='clj-long'>1218946708</span>","value":"1218946708"},{"type":"html","content":"<span class='clj-long'>1204599257</span>","value":"1204599257"},{"type":"html","content":"<span class='clj-long'>1202440744</span>","value":"1202440744"},{"type":"html","content":"<span class='clj-long'>1069773626</span>","value":"1069773626"},{"type":"html","content":"<span class='clj-long'>1009893576</span>","value":"1009893576"},{"type":"html","content":"<span class='clj-long'>1012411211</span>","value":"1012411211"},{"type":"html","content":"<span class='clj-long'>1010887185</span>","value":"1010887185"},{"type":"html","content":"<span class='clj-long'>1010030357</span>","value":"1010030357"},{"type":"html","content":"<span class='clj-long'>1007370044</span>","value":"1007370044"},{"type":"html","content":"<span class='clj-long'>1007655148</span>","value":"1007655148"},{"type":"html","content":"<span class='clj-long'>1150961025</span>","value":"1150961025"},{"type":"html","content":"<span class='clj-long'>1218434317</span>","value":"1218434317"},{"type":"html","content":"<span class='clj-long'>1004650772</span>","value":"1004650772"},{"type":"html","content":"<span class='clj-long'>1006944257</span>","value":"1006944257"},{"type":"html","content":"<span class='clj-long'>1004671165</span>","value":"1004671165"},{"type":"html","content":"<span class='clj-long'>1004784620</span>","value":"1004784620"},{"type":"html","content":"<span class='clj-long'>1015098268</span>","value":"1015098268"},{"type":"html","content":"<span class='clj-long'>1005502757</span>","value":"1005502757"},{"type":"html","content":"<span class='clj-long'>1007168766</span>","value":"1007168766"},{"type":"html","content":"<span class='clj-long'>1110257357</span>","value":"1110257357"},{"type":"html","content":"<span class='clj-long'>1017157529</span>","value":"1017157529"},{"type":"html","content":"<span class='clj-long'>1003904006</span>","value":"1003904006"},{"type":"html","content":"<span class='clj-long'>999619618</span>","value":"999619618"},{"type":"html","content":"<span class='clj-long'>1003459877</span>","value":"1003459877"},{"type":"html","content":"<span class='clj-long'>1017622695</span>","value":"1017622695"},{"type":"html","content":"<span class='clj-long'>1007235538</span>","value":"1007235538"},{"type":"html","content":"<span class='clj-long'>1004257438</span>","value":"1004257438"},{"type":"html","content":"<span class='clj-long'>1003050662</span>","value":"1003050662"},{"type":"html","content":"<span class='clj-long'>1007722591</span>","value":"1007722591"},{"type":"html","content":"<span class='clj-long'>1004372313</span>","value":"1004372313"},{"type":"html","content":"<span class='clj-long'>1009445503</span>","value":"1009445503"},{"type":"html","content":"<span class='clj-long'>1006792103</span>","value":"1006792103"},{"type":"html","content":"<span class='clj-long'>1004114848</span>","value":"1004114848"},{"type":"html","content":"<span class='clj-long'>1005288698</span>","value":"1005288698"},{"type":"html","content":"<span class='clj-long'>1008897210</span>","value":"1008897210"},{"type":"html","content":"<span class='clj-long'>1005974211</span>","value":"1005974211"},{"type":"html","content":"<span class='clj-long'>1005982395</span>","value":"1005982395"},{"type":"html","content":"<span class='clj-long'>1010890041</span>","value":"1010890041"}],"value":"(1008929870 1006691623 1007274197 1005775539 1009803490 1014054834 1008616404 1006501909 1014588731 1042267135 1005476415 1007543463 1009080563 1006871413 1006107920 1005764034 1007602449 1007358811 1010391196 1009402876 1010927691 1041252800 1218946708 1204599257 1202440744 1069773626 1009893576 1012411211 1010887185 1010030357 1007370044 1007655148 1150961025 1218434317 1004650772 1006944257 1004671165 1004784620 1015098268 1005502757 1007168766 1110257357 1017157529 1003904006 999619618 1003459877 1017622695 1007235538 1004257438 1003050662 1007722591 1004372313 1009445503 1006792103 1004114848 1005288698 1008897210 1005974211 1005982395 1010890041)"}],"value":"[:samples (1008929870 1006691623 1007274197 1005775539 1009803490 1014054834 1008616404 1006501909 1014588731 1042267135 1005476415 1007543463 1009080563 1006871413 1006107920 1005764034 1007602449 1007358811 1010391196 1009402876 1010927691 1041252800 1218946708 1204599257 1202440744 1069773626 1009893576 1012411211 1010887185 1010030357 1007370044 1007655148 1150961025 1218434317 1004650772 1006944257 1004671165 1004784620 1015098268 1005502757 1007168766 1110257357 1017157529 1003904006 999619618 1003459877 1017622695 1007235538 1004257438 1003050662 1007722591 1004372313 1009445503 1006792103 1004114848 1005288698 1008897210 1005974211 1005982395 1010890041)]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:runtime-details</span>","value":":runtime-details"},{"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:spec-vendor</span>","value":":spec-vendor"},{"type":"html","content":"<span class='clj-string'>&quot;Oracle Corporation&quot;</span>","value":"\"Oracle Corporation\""}],"value":"[:spec-vendor \"Oracle Corporation\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:spec-name</span>","value":":spec-name"},{"type":"html","content":"<span class='clj-string'>&quot;Java Virtual Machine Specification&quot;</span>","value":"\"Java Virtual Machine Specification\""}],"value":"[:spec-name \"Java Virtual Machine Specification\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:vm-version</span>","value":":vm-version"},{"type":"html","content":"<span class='clj-string'>&quot;24.45-b08&quot;</span>","value":"\"24.45-b08\""}],"value":"[:vm-version \"24.45-b08\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:name</span>","value":":name"},{"type":"html","content":"<span class='clj-string'>&quot;22855@rui-linux&quot;</span>","value":"\"22855@rui-linux\""}],"value":"[:name \"22855@rui-linux\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:clojure-version-string</span>","value":":clojure-version-string"},{"type":"html","content":"<span class='clj-string'>&quot;1.6.0&quot;</span>","value":"\"1.6.0\""}],"value":"[:clojure-version-string \"1.6.0\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:java-runtime-version</span>","value":":java-runtime-version"},{"type":"html","content":"<span class='clj-string'>&quot;1.7.0_45-b18&quot;</span>","value":"\"1.7.0_45-b18\""}],"value":"[:java-runtime-version \"1.7.0_45-b18\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:java-version</span>","value":":java-version"},{"type":"html","content":"<span class='clj-string'>&quot;1.7.0_45&quot;</span>","value":"\"1.7.0_45\""}],"value":"[:java-version \"1.7.0_45\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:vm-name</span>","value":":vm-name"},{"type":"html","content":"<span class='clj-string'>&quot;Java HotSpot(TM) 64-Bit Server VM&quot;</span>","value":"\"Java HotSpot(TM) 64-Bit Server VM\""}],"value":"[:vm-name \"Java HotSpot(TM) 64-Bit Server VM\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:vm-vendor</span>","value":":vm-vendor"},{"type":"html","content":"<span class='clj-string'>&quot;Oracle Corporation&quot;</span>","value":"\"Oracle Corporation\""}],"value":"[:vm-vendor \"Oracle Corporation\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:clojure-version</span>","value":":clojure-version"},{"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:major</span>","value":":major"},{"type":"html","content":"<span class='clj-unkown'>1</span>","value":"1"}],"value":"[:major 1]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:minor</span>","value":":minor"},{"type":"html","content":"<span class='clj-unkown'>6</span>","value":"6"}],"value":"[:minor 6]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:incremental</span>","value":":incremental"},{"type":"html","content":"<span class='clj-unkown'>0</span>","value":"0"}],"value":"[:incremental 0]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:qualifier</span>","value":":qualifier"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}],"value":"[:qualifier nil]"}],"value":"{:major 1, :minor 6, :incremental 0, :qualifier nil}"}],"value":"[:clojure-version {:major 1, :minor 6, :incremental 0, :qualifier nil}]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:spec-version</span>","value":":spec-version"},{"type":"html","content":"<span class='clj-string'>&quot;1.7&quot;</span>","value":"\"1.7\""}],"value":"[:spec-version \"1.7\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:sun-arch-data-model</span>","value":":sun-arch-data-model"},{"type":"html","content":"<span class='clj-string'>&quot;64&quot;</span>","value":"\"64\""}],"value":"[:sun-arch-data-model \"64\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:input-arguments</span>","value":":input-arguments"},{"type":"list-like","open":"<span class='clj-vector'>[<span>","close":"<span class='clj-vector'>]</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-string'>&quot;-Dclojure.compile.path=/home/rui/git/mine/utils/mongodb/log-analyzer/target/classes&quot;</span>","value":"\"-Dclojure.compile.path=/home/rui/git/mine/utils/mongodb/log-analyzer/target/classes\""},{"type":"html","content":"<span class='clj-string'>&quot;-Ddefault.version=0.1.0-SNAPSHOT&quot;</span>","value":"\"-Ddefault.version=0.1.0-SNAPSHOT\""},{"type":"html","content":"<span class='clj-string'>&quot;-Dfile.encoding=UTF-8&quot;</span>","value":"\"-Dfile.encoding=UTF-8\""},{"type":"html","content":"<span class='clj-string'>&quot;-Dclojure.debug=false&quot;</span>","value":"\"-Dclojure.debug=false\""}],"value":"[\"-Dclojure.compile.path=/home/rui/git/mine/utils/mongodb/log-analyzer/target/classes\" \"-Ddefault.version=0.1.0-SNAPSHOT\" \"-Dfile.encoding=UTF-8\" \"-Dclojure.debug=false\"]"}],"value":"[:input-arguments [\"-Dclojure.compile.path=/home/rui/git/mine/utils/mongodb/log-analyzer/target/classes\" \"-Ddefault.version=0.1.0-SNAPSHOT\" \"-Dfile.encoding=UTF-8\" \"-Dclojure.debug=false\"]]"}],"value":"{:spec-vendor \"Oracle Corporation\", :spec-name \"Java Virtual Machine Specification\", :vm-version \"24.45-b08\", :name \"22855@rui-linux\", :clojure-version-string \"1.6.0\", :java-runtime-version \"1.7.0_45-b18\", :java-version \"1.7.0_45\", :vm-name \"Java HotSpot(TM) 64-Bit Server VM\", :vm-vendor \"Oracle Corporation\", :clojure-version {:major 1, :minor 6, :incremental 0, :qualifier nil}, :spec-version \"1.7\", :sun-arch-data-model \"64\", :input-arguments [\"-Dclojure.compile.path=/home/rui/git/mine/utils/mongodb/log-analyzer/target/classes\" \"-Ddefault.version=0.1.0-SNAPSHOT\" \"-Dfile.encoding=UTF-8\" \"-Dclojure.debug=false\"]}"}],"value":"[:runtime-details {:spec-vendor \"Oracle Corporation\", :spec-name \"Java Virtual Machine Specification\", :vm-version \"24.45-b08\", :name \"22855@rui-linux\", :clojure-version-string \"1.6.0\", :java-runtime-version \"1.7.0_45-b18\", :java-version \"1.7.0_45\", :vm-name \"Java HotSpot(TM) 64-Bit Server VM\", :vm-vendor \"Oracle Corporation\", :clojure-version {:major 1, :minor 6, :incremental 0, :qualifier nil}, :spec-version \"1.7\", :sun-arch-data-model \"64\", :input-arguments [\"-Dclojure.compile.path=/home/rui/git/mine/utils/mongodb/log-analyzer/target/classes\" \"-Ddefault.version=0.1.0-SNAPSHOT\" \"-Dfile.encoding=UTF-8\" \"-Dclojure.debug=false\"]}]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:mean</span>","value":":mean"},{"type":"list-like","open":"<span class='clj-vector'>[<span>","close":"<span class='clj-vector'>]</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>5.999796657523588E-5</span>","value":"5.999796657523588E-5"},{"type":"list-like","open":"<span class='clj-lazy-seq'>(<span>","close":"<span class='clj-lazy-seq'>)</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>5.934882563855657E-5</span>","value":"5.934882563855657E-5"},{"type":"html","content":"<span class='clj-double'>6.1006796081120524E-5</span>","value":"6.1006796081120524E-5"}],"value":"(5.934882563855657E-5 6.1006796081120524E-5)"}],"value":"[5.999796657523588E-5 (5.934882563855657E-5 6.1006796081120524E-5)]"}],"value":"[:mean [5.999796657523588E-5 (5.934882563855657E-5 6.1006796081120524E-5)]]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:final-gc-time</span>","value":":final-gc-time"},{"type":"html","content":"<span class='clj-long'>71633004</span>","value":"71633004"}],"value":"[:final-gc-time 71633004]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:execution-count</span>","value":":execution-count"},{"type":"html","content":"<span class='clj-long'>17135</span>","value":"17135"}],"value":"[:execution-count 17135]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:variance</span>","value":":variance"},{"type":"list-like","open":"<span class='clj-vector'>[<span>","close":"<span class='clj-vector'>]</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>1.0878639113178648E-11</span>","value":"1.0878639113178648E-11"},{"type":"list-like","open":"<span class='clj-lazy-seq'>(<span>","close":"<span class='clj-lazy-seq'>)</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>4.600684641387765E-12</span>","value":"4.600684641387765E-12"},{"type":"html","content":"<span class='clj-double'>2.0588486965645227E-11</span>","value":"2.0588486965645227E-11"}],"value":"(4.600684641387765E-12 2.0588486965645227E-11)"}],"value":"[1.0878639113178648E-11 (4.600684641387765E-12 2.0588486965645227E-11)]"}],"value":"[:variance [1.0878639113178648E-11 (4.600684641387765E-12 2.0588486965645227E-11)]]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:os-details</span>","value":":os-details"},{"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:arch</span>","value":":arch"},{"type":"html","content":"<span class='clj-string'>&quot;amd64&quot;</span>","value":"\"amd64\""}],"value":"[:arch \"amd64\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:available-processors</span>","value":":available-processors"},{"type":"html","content":"<span class='clj-unkown'>12</span>","value":"12"}],"value":"[:available-processors 12]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:name</span>","value":":name"},{"type":"html","content":"<span class='clj-string'>&quot;Linux&quot;</span>","value":"\"Linux\""}],"value":"[:name \"Linux\"]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:version</span>","value":":version"},{"type":"html","content":"<span class='clj-string'>&quot;3.13-1-amd64&quot;</span>","value":"\"3.13-1-amd64\""}],"value":"[:version \"3.13-1-amd64\"]"}],"value":"{:arch \"amd64\", :available-processors 12, :name \"Linux\", :version \"3.13-1-amd64\"}"}],"value":"[:os-details {:arch \"amd64\", :available-processors 12, :name \"Linux\", :version \"3.13-1-amd64\"}]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:tail-quantile</span>","value":":tail-quantile"},{"type":"html","content":"<span class='clj-double'>0.025</span>","value":"0.025"}],"value":"[:tail-quantile 0.025]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:outlier-variance</span>","value":":outlier-variance"},{"type":"html","content":"<span class='clj-double'>0.4017062122270068</span>","value":"0.4017062122270068"}],"value":"[:outlier-variance 0.4017062122270068]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:outliers</span>","value":":outliers"},{"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:low-severe</span>","value":":low-severe"},{"type":"html","content":"<span class='clj-long'>0</span>","value":"0"}],"value":"[:low-severe 0]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:low-mild</span>","value":":low-mild"},{"type":"html","content":"<span class='clj-long'>0</span>","value":"0"}],"value":"[:low-mild 0]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:high-mild</span>","value":":high-mild"},{"type":"html","content":"<span class='clj-long'>0</span>","value":"0"}],"value":"[:high-mild 0]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:high-severe</span>","value":":high-severe"},{"type":"html","content":"<span class='clj-long'>9</span>","value":"9"}],"value":"[:high-severe 9]"}],"value":"#criterium.core.OutlierCount{:low-severe 0, :low-mild 0, :high-mild 0, :high-severe 9}"}],"value":"[:outliers #criterium.core.OutlierCount{:low-severe 0, :low-mild 0, :high-mild 0, :high-severe 9}]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:warmup-time</span>","value":":warmup-time"},{"type":"html","content":"<span class='clj-long'>10907123551</span>","value":"10907123551"}],"value":"[:warmup-time 10907123551]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:lower-q</span>","value":":lower-q"},{"type":"list-like","open":"<span class='clj-vector'>[<span>","close":"<span class='clj-vector'>]</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>5.853811858768603E-5</span>","value":"5.853811858768603E-5"},{"type":"list-like","open":"<span class='clj-lazy-seq'>(<span>","close":"<span class='clj-lazy-seq'>)</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>5.8337882579515616E-5</span>","value":"5.8337882579515616E-5"},{"type":"html","content":"<span class='clj-double'>5.859376457251241E-5</span>","value":"5.859376457251241E-5"}],"value":"(5.8337882579515616E-5 5.859376457251241E-5)"}],"value":"[5.853811858768603E-5 (5.8337882579515616E-5 5.859376457251241E-5)]"}],"value":"[:lower-q [5.853811858768603E-5 (5.8337882579515616E-5 5.859376457251241E-5)]]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:warmup-executions</span>","value":":warmup-executions"},{"type":"html","content":"<span class='clj-long'>185743</span>","value":"185743"}],"value":"[:warmup-executions 185743]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:sample-count</span>","value":":sample-count"},{"type":"html","content":"<span class='clj-long'>60</span>","value":"60"}],"value":"[:sample-count 60]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:upper-q</span>","value":":upper-q"},{"type":"list-like","open":"<span class='clj-vector'>[<span>","close":"<span class='clj-vector'>]</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>7.07244040560257E-5</span>","value":"7.07244040560257E-5"},{"type":"list-like","open":"<span class='clj-lazy-seq'>(<span>","close":"<span class='clj-lazy-seq'>)</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>6.60418314969361E-5</span>","value":"6.60418314969361E-5"},{"type":"html","content":"<span class='clj-double'>7.113782947184128E-5</span>","value":"7.113782947184128E-5"}],"value":"(6.60418314969361E-5 7.113782947184128E-5)"}],"value":"[7.07244040560257E-5 (6.60418314969361E-5 7.113782947184128E-5)]"}],"value":"[:upper-q [7.07244040560257E-5 (6.60418314969361E-5 7.113782947184128E-5)]]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:total-time</span>","value":":total-time"},{"type":"html","content":"<span class='clj-double'>61.660551799</span>","value":"61.660551799"}],"value":"[:total-time 61.660551799]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:sample-variance</span>","value":":sample-variance"},{"type":"list-like","open":"<span class='clj-vector'>[<span>","close":"<span class='clj-vector'>]</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>1.0369914151231894E-11</span>","value":"1.0369914151231894E-11"},{"type":"list-like","open":"<span class='clj-lazy-seq'>(<span>","close":"<span class='clj-lazy-seq'>)</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>0.0</span>","value":"0.0"},{"type":"html","content":"<span class='clj-double'>0.0</span>","value":"0.0"}],"value":"(0.0 0.0)"}],"value":"[1.0369914151231894E-11 (0.0 0.0)]"}],"value":"[:sample-variance [1.0369914151231894E-11 (0.0 0.0)]]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:options</span>","value":":options"},{"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:overhead</span>","value":":overhead"},{"type":"html","content":"<span class='clj-double'>1.662542005421643E-9</span>","value":"1.662542005421643E-9"}],"value":"[:overhead 1.662542005421643E-9]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:max-gc-attempts</span>","value":":max-gc-attempts"},{"type":"html","content":"<span class='clj-long'>100</span>","value":"100"}],"value":"[:max-gc-attempts 100]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:samples</span>","value":":samples"},{"type":"html","content":"<span class='clj-long'>60</span>","value":"60"}],"value":"[:samples 60]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:target-execution-time</span>","value":":target-execution-time"},{"type":"html","content":"<span class='clj-long'>1000000000</span>","value":"1000000000"}],"value":"[:target-execution-time 1000000000]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:warmup-jit-period</span>","value":":warmup-jit-period"},{"type":"html","content":"<span class='clj-long'>10000000000</span>","value":"10000000000"}],"value":"[:warmup-jit-period 10000000000]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:tail-quantile</span>","value":":tail-quantile"},{"type":"html","content":"<span class='clj-double'>0.025</span>","value":"0.025"}],"value":"[:tail-quantile 0.025]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:bootstrap-size</span>","value":":bootstrap-size"},{"type":"html","content":"<span class='clj-long'>1000</span>","value":"1000"}],"value":"[:bootstrap-size 1000]"}],"value":"{:overhead 1.662542005421643E-9, :max-gc-attempts 100, :samples 60, :target-execution-time 1000000000, :warmup-jit-period 10000000000, :tail-quantile 0.025, :bootstrap-size 1000}"}],"value":"[:options {:overhead 1.662542005421643E-9, :max-gc-attempts 100, :samples 60, :target-execution-time 1000000000, :warmup-jit-period 10000000000, :tail-quantile 0.025, :bootstrap-size 1000}]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:sample-mean</span>","value":":sample-mean"},{"type":"list-like","open":"<span class='clj-vector'>[<span>","close":"<span class='clj-vector'>]</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>5.9975247348506966E-5</span>","value":"5.9975247348506966E-5"},{"type":"list-like","open":"<span class='clj-lazy-seq'>(<span>","close":"<span class='clj-lazy-seq'>)</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-double'>5.031454202430382E-5</span>","value":"5.031454202430382E-5"},{"type":"html","content":"<span class='clj-double'>6.963595267271011E-5</span>","value":"6.963595267271011E-5"}],"value":"(5.031454202430382E-5 6.963595267271011E-5)"}],"value":"[5.9975247348506966E-5 (5.031454202430382E-5 6.963595267271011E-5)]"}],"value":"[:sample-mean [5.9975247348506966E-5 (5.031454202430382E-5 6.963595267271011E-5)]]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:results</span>","value":":results"},{"type":"list-like","open":"<span class='clj-lazy-seq'>(<span>","close":"<span class='clj-lazy-seq'>)</span>","separator":" ","items":[{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"},{"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}],"value":"(nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil)"}],"value":"[:results (nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil)]"}],"value":"{:overhead 1.662542005421643E-9, :samples (1008929870 1006691623 1007274197 1005775539 1009803490 1014054834 1008616404 1006501909 1014588731 1042267135 1005476415 1007543463 1009080563 1006871413 1006107920 1005764034 1007602449 1007358811 1010391196 1009402876 1010927691 1041252800 1218946708 1204599257 1202440744 1069773626 1009893576 1012411211 1010887185 1010030357 1007370044 1007655148 1150961025 1218434317 1004650772 1006944257 1004671165 1004784620 1015098268 1005502757 1007168766 1110257357 1017157529 1003904006 999619618 1003459877 1017622695 1007235538 1004257438 1003050662 1007722591 1004372313 1009445503 1006792103 1004114848 1005288698 1008897210 1005974211 1005982395 1010890041), :runtime-details {:spec-vendor \"Oracle Corporation\", :spec-name \"Java Virtual Machine Specification\", :vm-version \"24.45-b08\", :name \"22855@rui-linux\", :clojure-version-string \"1.6.0\", :java-runtime-version \"1.7.0_45-b18\", :java-version \"1.7.0_45\", :vm-name \"Java HotSpot(TM) 64-Bit Server VM\", :vm-vendor \"Oracle Corporation\", :clojure-version {:major 1, :minor 6, :incremental 0, :qualifier nil}, :spec-version \"1.7\", :sun-arch-data-model \"64\", :input-arguments [\"-Dclojure.compile.path=/home/rui/git/mine/utils/mongodb/log-analyzer/target/classes\" \"-Ddefault.version=0.1.0-SNAPSHOT\" \"-Dfile.encoding=UTF-8\" \"-Dclojure.debug=false\"]}, :mean [5.999796657523588E-5 (5.934882563855657E-5 6.1006796081120524E-5)], :final-gc-time 71633004, :execution-count 17135, :variance [1.0878639113178648E-11 (4.600684641387765E-12 2.0588486965645227E-11)], :os-details {:arch \"amd64\", :available-processors 12, :name \"Linux\", :version \"3.13-1-amd64\"}, :tail-quantile 0.025, :outlier-variance 0.4017062122270068, :outliers #criterium.core.OutlierCount{:low-severe 0, :low-mild 0, :high-mild 0, :high-severe 9}, :warmup-time 10907123551, :lower-q [5.853811858768603E-5 (5.8337882579515616E-5 5.859376457251241E-5)], :warmup-executions 185743, :sample-count 60, :upper-q [7.07244040560257E-5 (6.60418314969361E-5 7.113782947184128E-5)], :total-time 61.660551799, :sample-variance [1.0369914151231894E-11 (0.0 0.0)], :options {:overhead 1.662542005421643E-9, :max-gc-attempts 100, :samples 60, :target-execution-time 1000000000, :warmup-jit-period 10000000000, :tail-quantile 0.025, :bootstrap-size 1000}, :sample-mean [5.9975247348506966E-5 (5.031454202430382E-5 6.963595267271011E-5)], :results (nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil nil)}"}
;; <=

;; @@
(assoc results :5000 (to-us (take-mean re-5000)))
;; @@
;; =>
;;; {"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:5000</span>","value":":5000"},{"type":"html","content":"<span class='clj-double'>5.890269131143382</span>","value":"5.890269131143382"}],"value":"[:5000 5.890269131143382]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:10000</span>","value":":10000"},{"type":"html","content":"<span class='clj-double'>58.206438170025216</span>","value":"58.206438170025216"}],"value":"[:10000 58.206438170025216]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:50000</span>","value":":50000"},{"type":"html","content":"<span class='clj-double'>58.57084298408284</span>","value":"58.57084298408284"}],"value":"[:50000 58.57084298408284]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:100000</span>","value":":100000"},{"type":"html","content":"<span class='clj-double'>58.97364480124979</span>","value":"58.97364480124979"}],"value":"[:100000 58.97364480124979]"}],"value":"{:5000 5.890269131143382, :10000 58.206438170025216, :50000 58.57084298408284, :100000 58.97364480124979}"}
;; <=

;; @@
(def db (mg/get-db (mg/connect) "test"))

(def re-with-index
  (into {} 
        (for [i [1 1000 5000 10000 50000 100000 200000 400000 1000000]] 
          {(keyword (str i)) (run-test! {:a 10} i 1 db "foo")}))
  )
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/re-with-index</span>","value":"#'mongo-analyzer.gen/re-with-index"}
;; <=

;; @@
 
(def results-idx (into {} (for [[k, v] re-with-index]
  [k (* 1000000 (get (:mean v) 0))]
  )))
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/results-idx</span>","value":"#'mongo-analyzer.gen/results-idx"}
;; <=

;; @@
results-idx
;; @@
;; =>
;;; {"type":"list-like","open":"<span class='clj-map'>{<span>","close":"<span class='clj-map'>}</span>","separator":", ","items":[{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:10000</span>","value":":10000"},{"type":"html","content":"<span class='clj-double'>59.74729581780057</span>","value":"59.74729581780057"}],"value":"[:10000 59.74729581780057]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:50000</span>","value":":50000"},{"type":"html","content":"<span class='clj-double'>59.29842581398064</span>","value":"59.29842581398064"}],"value":"[:50000 59.29842581398064]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:5000</span>","value":":5000"},{"type":"html","content":"<span class='clj-double'>60.19394025002922</span>","value":"60.19394025002922"}],"value":"[:5000 60.19394025002922]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:1000</span>","value":":1000"},{"type":"html","content":"<span class='clj-double'>59.840537183679835</span>","value":"59.840537183679835"}],"value":"[:1000 59.840537183679835]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:400000</span>","value":":400000"},{"type":"html","content":"<span class='clj-double'>59.43426550189118</span>","value":"59.43426550189118"}],"value":"[:400000 59.43426550189118]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:1</span>","value":":1"},{"type":"html","content":"<span class='clj-double'>58.885749301771455</span>","value":"58.885749301771455"}],"value":"[:1 58.885749301771455]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:100000</span>","value":":100000"},{"type":"html","content":"<span class='clj-double'>58.88137562339195</span>","value":"58.88137562339195"}],"value":"[:100000 58.88137562339195]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:200000</span>","value":":200000"},{"type":"html","content":"<span class='clj-double'>59.86533383341102</span>","value":"59.86533383341102"}],"value":"[:200000 59.86533383341102]"},{"type":"list-like","open":"","close":"","separator":" ","items":[{"type":"html","content":"<span class='clj-keyword'>:1000000</span>","value":":1000000"},{"type":"html","content":"<span class='clj-double'>58.93977570936479</span>","value":"58.93977570936479"}],"value":"[:1000000 58.93977570936479]"}],"value":"{:10000 59.74729581780057, :50000 59.29842581398064, :5000 60.19394025002922, :1000 59.840537183679835, :400000 59.43426550189118, :1 58.885749301771455, :100000 58.88137562339195, :200000 59.86533383341102, :1000000 58.93977570936479}"}
;; <=

;; @@
;; 2.4.9

(def db (mg/get-db (mg/connect) "test"))

(def re-with-index-249
  (into {} 
        (for [i [1 1000 5000 10000 50000 100000 200000 400000 1000000]] 
          {(keyword (str i)) (run-test! {:a 10} i 1 db "foo")}))
  )
 
(def results-idx-249 (into {} (for [[k, v] re-with-index-249]
  [k (* 1000000 (get (:mean v) 0))]
  )))
;; @@

;; **
;;; ## Query -> Perf	
;; **

;; @@
(def query-1 {:q {:OP_query {:status "OPEN" :cid "ObjectId('538fa43177143b548c76ecfd')" :cre {:OP_lte "new-Date(1401923435593)"}} :OP_orderby {:cre "-1"}} 
              :query-plan "IXSCAN"})
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/query-1</span>","value":"#'mongo-analyzer.gen/query-1"}
;; <=

;; @@
(p/pprint query-1)
;; @@
;; ->
;;; {:q
;;;  {:OP_query
;;;   {:cid &quot;ObjectId(&#x27;538fa43177143b548c76ecfd&#x27;)&quot;,
;;;    :cre {:OP_lte &quot;new-Date(1401923435593)&quot;},
;;;    :status &quot;OPEN&quot;},
;;;   :OP_orderby {:cre &quot;-1&quot;}},
;;;  :query-plan &quot;IXSCAN&quot;}
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; @@
(defn query->schema 
  "generate data schema from query"
  [q]
  (let [query (:OP_query (:q q))] 
    (into {} (for [ [k v] query]
               (cond 
                (instance? String v)  
                  (cond 
                   (re-matches #"ObjectId.*" v)
                   {k :ObjectId}
                   :else
                   {k :string})
                (contains? v :OP_lte) {k [:rand-int :lte]}
                :else {k v}
                )))))
  
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/query-&gt;schema</span>","value":"#'mongo-analyzer.gen/query->schema"}
;; <=

;; @@
(p/pprint query-1) (prn)
(p/pprint (query->schema query-1)) (prn)

(p/pprint (gen-docs  (query->schema query-1) 10 10 :query false))
;; @@
;; ->
;;; {:q
;;;  {:OP_query
;;;   {:cid &quot;ObjectId(&#x27;538fa43177143b548c76ecfd&#x27;)&quot;,
;;;    :cre {:OP_lte &quot;new-Date(1401923435593)&quot;},
;;;    :status &quot;OPEN&quot;},
;;;   :OP_orderby {:cre &quot;-1&quot;}},
;;;  :query-plan &quot;IXSCAN&quot;}
;;; 
;;; {:cid :ObjectId, :cre [:rand-int :lte], :status :string}
;;; 
;;; ({:_id 10, :cid &quot;ObjectId(10)&quot;, :cre 516, :status &quot;string_10&quot;}
;;;  {:_id 11, :cid &quot;ObjectId(11)&quot;, :cre 442, :status &quot;string_11&quot;}
;;;  {:_id 12, :cid &quot;ObjectId(12)&quot;, :cre 967, :status &quot;string_12&quot;}
;;;  {:_id 13, :cid &quot;ObjectId(13)&quot;, :cre 322, :status &quot;string_13&quot;}
;;;  {:_id 14, :cid &quot;ObjectId(14)&quot;, :cre 563, :status &quot;string_14&quot;}
;;;  {:_id 15, :cid &quot;ObjectId(15)&quot;, :cre 894, :status &quot;string_15&quot;}
;;;  {:_id 16, :cid &quot;ObjectId(16)&quot;, :cre 540, :status &quot;string_16&quot;}
;;;  {:_id 17, :cid &quot;ObjectId(17)&quot;, :cre 326, :status &quot;string_17&quot;}
;;;  {:_id 18, :cid &quot;ObjectId(18)&quot;, :cre 551, :status &quot;string_18&quot;}
;;;  {:_id 19, :cid &quot;ObjectId(19)&quot;, :cre 894, :status &quot;string_19&quot;})
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; @@
(def db (mg/get-db (mg/connect) "test"))
(def coll "test")

;;(run-test! (query->schema query-1) 5000 1 db coll)

(def result-1 (run-test! (query->schema query-1) 5000 1 db coll))
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/result-1</span>","value":"#'mongo-analyzer.gen/result-1"}
;; <=

;; @@
(report-result result-1)
;; @@
;; ->
;;; Evaluation count : 602160 in 60 samples of 10036 calls.
;;;              Execution time mean : 99.885183 µs
;;;     Execution time std-deviation : 2.282672 µs
;;;    Execution time lower quantile : 98.803143 µs ( 2.5%)
;;;    Execution time upper quantile : 106.053833 µs (97.5%)
;;;                    Overhead used : 1.679195 ns
;;; 
;;; Found 5 outliers in 60 samples (8.3333 %)
;;; 	low-severe	 1 (1.6667 %)
;;; 	low-mild	 4 (6.6667 %)
;;;  Variance from outliers : 10.9880 % Variance is moderately inflated by outliers
;;; 
;; <-
;; =>
;;; {"type":"html","content":"<span class='clj-nil'>nil</span>","value":"nil"}
;; <=

;; @@
(def result-1(run-test! (query->schema query-1) 50000 1 db coll))
;; @@
;; =>
;;; {"type":"html","content":"<span class='clj-var'>#&#x27;mongo-analyzer.gen/result-1</span>","value":"#'mongo-analyzer.gen/result-1"}
;; <=

;; @@

;; @@
