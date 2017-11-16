# Helm charts for examples

We have copies of some Helm charts from different locations when we need to do
some customizations of them that helps us making them working out of the box for
our examples. Usual changes:

* NodePort as Service types by default
* imagePullPolicy: IfNotPresent for all images
* enabling some behaviours by default

List of charts and their initial locations (with repo git sha if available)

* Istio
    * Version: 0.2.7-chart6
    * Repo: https://github.com/kubernetes/charts
    * Commit: 65181acc03ea03fb47d264d03f7ea3ffd2aa32c7
    * License: Apache License 2.0

* HDFS, Kafka, Spark, TweePub, TweeTics, TweeViz, ZooKeeper
    * Repo: https://github.com/Mirantis/k8s-apps
    * Commit: Ic0653fa8351b332b8987f37aa10461223e009c33
    * License: Apache License 2.0
