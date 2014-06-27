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
(use 'mongo-analyzer.log-analyzer)
(use 'mongo-analyzer.utils)

;; read log and convert to map
(defn normalize-string [s]
  (if (= nil s)
    ""
    (->> s
         (str-replace #"\"" ""))))

(defn get-mms-agent-id-from-url [u]
  (get (re-find #"^/ping/v[12]/(.+)$" u) 1))

(defn mapping-mms-access-log [l]
  (let [t (split l #"\s")]
    {
     :ip (normalize-string (get t 0))
     :ts (str-replace #"\[" "" (get t 3))
     :op (str-replace #"\"" "" (get t 5))
     :url(get t 6)
     :http-agent (normalize-string (get t 11))
     :http (str (str-replace #"\"" "" (get t 5)) " " (get t 6))
     :mms-agent-id (get-mms-agent-id-from-url (get t 6))
;     :o_ip_str (split (normalize-string (get t 14)) #",")
;     :o_ip (let [i  (get t 14)]
;             (if (= "" i)
;               "n/a"
;               (get (split (str-replace #"\"" "" i) #",") 4)
;             ))
     }
    ))


(defn parse-http-log [file-name]
  (map mapping-mms-access-log (get-lines file-name)))

(defn unique-http-client [d]
  (distinct (filter #(not= nil %) (map #(:mms-agent-id %) d))))

;; test
;(def data (parse-http-log "/Users/rui/work/mms-load/prod/mms-prod-http-log-10k.log"))
(def data (parse-http-log "/Users/rui/work/mms-load/prod/mms-prod-http-log-all.log"))

(def data (parse-http-log "/Users/rui/work/mms-load/prod/t"))
(take 1 (filter #(not-nil? (:mms-agent-id %)) data))



(count (unique-http-client data))
