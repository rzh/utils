(ns mongo-analyzer.utils
  (:require [clojure.string :as string])
  (:import (java.util.concurrent Executors
                                 ScheduledThreadPoolExecutor
                                 ThreadPoolExecutor
                                 SynchronousQueue
                                 TimeUnit)))


(use 'clojure.java.io)

;; utils

;; util functions
(def not-nil? (complement nil?))

(defn str-replace [pattern replacement string]
  (if (= nil string)
    ""
    (string/replace string pattern replacement)))

(defn get-lines [fname]
  (with-open [r (reader fname)]
    (doall (line-seq r))))


;; ++++++++++++++++++++++++++++
;; load utils
(defn wrap-latency
  "Returns a function which takes an argument and calls (f arg). (f arg) should
  return a map. Wrap-time will add a new key :latency to this map, which is the
  number of milliseconds the call took."
  [f]
  (fn measure-latency [req]
    (let [t1 (System/nanoTime)
          v  (f req)
          t2 (System/nanoTime)]
      (assoc v :latency (/ (double (- t2 t1)) 1000000.0)))))


(defn run-test! [niters]
  (let [nitems   10
        nthreads 10
        refs  (map ref (repeat nitems 0))
        pool  (Executors/newFixedThreadPool nthreads)
        tasks (map (fn [t]
                      (fn []
                        (dotimes [n niters]
                              (* n n))))
                   (range nthreads))]
    (doseq [future (.invokeAll pool tasks)]
      (.get future))
    (.shutdown pool)
    {:result (map deref refs)}))



;; ----------
(def ^:dynamic *v*)

(defn incv [n] (set! *v* (+ *v* n)))
(defn test-vars [niters]
  (let [nthreads 10
        pool (Executors/newFixedThreadPool nthreads)
        tasks (map (fn [t]
                     #(binding [*v* 0]
                        (dotimes [n niters]
                          (incv t))
                        *v*))
                   (range nthreads))]
      (let [ret (.invokeAll pool tasks)]
        (.shutdown pool)
        {:result (map #(.get %) ret)})))




;; ++++++++++++++++++
((wrap-latency run-test!) 100000)

((wrap-latency test-vars) 100000)

(reduce + [1 2 3 4])
