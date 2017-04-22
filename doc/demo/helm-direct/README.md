# Direct Helm demo without Aptomi

* Make sure `kubectl` installed and configured to the needed cluster.

* Make sure `helm` installed.

* Run `./deploy.sh <namespace>` (for example, in namespace demo) to deploy demo
services to the specified namespace. It'll output helm logs and list of
services, such as:

```shell
NAME                    CLUSTER-IP   EXTERNAL-IP   PORT(S)                      AGE
hdfs-namenode           None         <none>        8020/TCP                     0s
hdfs-ui                 10.0.0.178   <nodes>       50070:30386/TCP              0s
kafka-kafka-1           None         <none>        9092/TCP                     2s
spark-1-spark-master    10.0.0.55    <none>        7077/TCP                     0s
spark-1-spark-restapi   10.0.0.29    <none>        6066/TCP                     0s
spark-1-spark-webui     10.0.0.163   <nodes>       8080:31000/TCP               0s
spark-1-zeppelin        10.0.0.180   <nodes>       8080:31001/TCP               0s
zk-kafka-1              None         <none>        2888/TCP,3888/TCP,2181/TCP   2s
```

Ports 3xxxx are NodePorts and could be used together with any IP of any K8s
cluster node. To get node IP for minikube run `minikube ip`.

* `helm status kafka-1` (or spark-1, hdfs-1) could be used to get full status
of the deployed chart. 

* Wait until all pods will move to `READY: 1/1 STATUS: Running` and Kafka/Spark
will get needed number of replicas deployed, you can use following command to
watch how states are changing:

```shell
> kubectl -n demo get po -w

NAME                                    READY     STATUS              RESTARTS   AGE
hdfs-datanode-351365524-50v62           0/1       Init:0/1            0          11s
hdfs-datanode-351365524-8n1nq           0/1       Init:0/1            0          11s
hdfs-datanode-351365524-s500h           0/1       Init:0/1            0          11s
hdfs-namenode-0                         0/1       Pending             0          11s
kafka-kafka-1-0                         0/1       Init:0/1            0          15s
spark-1-spark-master-569428674-m1ckc    0/1       ContainerCreating   0          14s
spark-1-spark-worker-3575640858-42l92   0/1       Pending             0          14s
spark-1-spark-worker-3575640858-99xl5   0/1       Terminating         0          10m
spark-1-spark-worker-3575640858-g2vbj   0/1       ContainerCreating   0          14s
spark-1-spark-worker-3575640858-gn5vr   0/1       Pending             0          14s
spark-1-spark-worker-3575640858-kqk0v   0/1       Terminating         0          10m
spark-1-spark-worker-3575640858-pndg7   0/1       Pending             0          14s
spark-1-spark-worker-3575640858-r2k8p   0/1       ContainerCreating   0          14s
spark-1-zeppelin-544894360-frq0v        0/1       ContainerCreating   0          14s
zk-kafka-1-0                            0/1       Running             0          15s
```

Finally it should looks like:

```shell
> kubectl -n demo get po

NAME                                    READY     STATUS    RESTARTS   AGE
hdfs-datanode-351365524-50v62           1/1       Running   0          2m
hdfs-datanode-351365524-8n1nq           1/1       Running   0          2m
hdfs-datanode-351365524-s500h           1/1       Running   0          2m
hdfs-namenode-0                         1/1       Running   0          2m
kafka-kafka-1-0                         1/1       Running   0          2m
kafka-kafka-1-1                         1/1       Running   0          1m
kafka-kafka-1-2                         1/1       Running   0          39s
spark-1-spark-master-569428674-m1ckc    1/1       Running   0          2m
spark-1-spark-worker-3575640858-42l92   1/1       Running   0          2m
spark-1-spark-worker-3575640858-g2vbj   1/1       Running   0          2m
spark-1-spark-worker-3575640858-gn5vr   1/1       Running   0          2m
spark-1-spark-worker-3575640858-pndg7   1/1       Running   0          2m
spark-1-spark-worker-3575640858-r2k8p   1/1       Running   0          2m
spark-1-zeppelin-544894360-frq0v        1/1       Running   0          2m
zk-kafka-1-0                            1/1       Running   0          2m
zk-kafka-1-1                            1/1       Running   0          2m
zk-kafka-1-2                            1/1       Running   0          1m
```

* To delete installed services use `helm --purge delete kafka-1 spark-1 hdfs-1`,
it'll cleanup resources. If something happened during deployment or you want to
make sure that everything deleted, run:
`kubectl delete ns demo && helm delete --purge kafka-1 spark-1 hdfs-1`.
