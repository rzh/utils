(defproject default "0.1.0-SNAPSHOT"
  :description "Log analyzer for mongoDB"
  :url "http://mongodb.org"
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :dependencies [[org.clojure/clojure "1.6.0"]
                 [org.clojure/data.json "0.2.4"]
                 [org.clojure/core.match "0.2.1"]
                 [criterium "0.4.3"]
                 ;; [congomongo "0.4.4"]
                 [ring/ring-jetty-adapter "1.2.1"]
                 [compojure "1.1.6"]
                 [hiccup "1.0.4"]
                 [http-kit "2.1.16"] 
                 [org.clojure/test.check "0.5.8"]
                 [com.novemberain/monger "2.0.0"]]
  :jvm-opts ^:replace ["-server"]
  :plugins [[lein-gorilla "0.2.0"]])
