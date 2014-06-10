(defproject default "0.1.0-SNAPSHOT"
  :description "Log analyzer for mongoDB"
  :url "http://mongodb.org"
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :dependencies [[org.clojure/clojure "1.5.1"]
                 [org.clojure/data.json "0.2.4"]
                 [org.clojure/core.match "0.2.1"]
                 [com.novemberain/monger "2.0.0-rc1"]]
  :plugins [[lein-gorilla "0.2.0"]])
