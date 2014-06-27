(ns mongo-analyzer.http-log-utils
  (:require [clojure.pprint :as p]
;            [clj-time.coerce :as c]
            [clojure.data.json :as json]
            [clojure.string :as string]
            ))

;(use 'gorilla-plot.core)
;(use 'gorilla-repl.table)
(use '[clojure.core.match :only (match)])

(use 'clojure.java.io)

(use '[clojure.string :only (join split)])
; (use 'mongo-analyzer.log-analyzer)
(use 'mongo-analyzer.utils)


;; to process op-log

(defn parse-oplog
  "Read and process oplog file,
  input is oplog file name.
  The log file shall be one JSON per line"
  [log-file]

  (map #(json/read-json %)(get-lines log-file))
  )


;; filters
(defn oplog-filter-insert-only [l]
  (= "i" (:op l)))

(defn oplog-filter-update-only [l]
  (= "u" (:op l)))


;; get all ns for oplog

(defn oplog-all-ns
  "input is array of maps of events"
  [d]
  (frequencies (map #(:ns %) d)))

;; filters
(defn oplog-filter-by-ns [n d]
  (= n (:ns d)))


;; test code

(def op-data (parse-oplog "/Users/rui/work/mms-load/oplog/mms-setup-monitor.log"))

(def all-insert (filter oplog-filter-insert-only op-data))

(p/pprint (take 2 (filter #(oplog-filter-by-ns "mmsdbminutes-even.rrdMinutes2014061318" %) all-insert)))
(p/pprint (take 2 (filter #(oplog-filter-by-ns "mmsdbminutes-odd.rrdMinutes2014061319" %) all-insert)))

(oplog-all-ns all-insert)


;; regular operations
(def op-data-reg (parse-oplog "/Users/rui/work/mms-load/oplog/mms-regular-one-host-monitoring.log"))
(count op-data-reg)

(p/pprint (->> op-data
               (filter oplog-filter-insert-only)
;               (filter #(oplog-filter-by-ns "mmsdbminutes-odd.rrdMinutes2014061319" %))
;               (filter #(oplog-filter-by-ns "mmsdbminutes-even.rrdMinutes2014061318" %))
               (filter #(oplog-filter-by-ns "mmsdbhours.data.rrdHours" %))
;               (filter #(= "2529457eab5e46b14d5aa4d99f7a83aa-d4e52a310654b047f46af98be05f8be3-2014061318-gcnum-maGlobalmemory" (:_id (:o %))))
;               (filter #(= "2529457eab5e46b14d5aa4d99f7a83aa-c2a6c770c4218571f6a6ac20a17e09c1-2014061318-delete-opcounters" (:_id (:o %))))
;               (map #(:_id (:o %)))
               (map #(str (:i (:o %)) "-" (:g (:o %))) )
               (distinct)
;               (sort)
               (count)
;               (take 1)
               ))


