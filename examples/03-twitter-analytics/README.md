# Demo scenario

In this example there are 2 main services:
- analytics_pipeline- it consists of kafka, spark, hdfs, zookeeper and is offered in two different contexts
  - *stage*: in sharing all consumers of analytics_pipeline get to share the same instance
  - *production*: a single production instance with more memory, better replicas count, etc 
- twitter_stats - it depends on analytics_pipeline and consists of 3 components (publisher, stats, ui).
  It gets data from Twitter in real time, calculates top hashtags using an external analytics_pipeline service and displays
  results on a web page.  

![Diagram](diagram.png)

These services have different owners, who can fully control how their services get offered and shared:
- Frank defined analytics_pipeline
- John defined twitter_stats
- Alice, Bob, Carol are services consumers and/or code contributors
- Sam is a domain admin for Aptomi

This example illustrates a few important things that Aptomi does:
1. Service definitions & instantiation
    - Service owners publish their services and fully define how they get offered & shares to others
    - Implementation details are abstracted away from consumers, so they can consume other services without knowing the inner-workings
1. Service update
    - Change parameter/label with a single command and re-apply only the required delta 
1. Service reuse
    - Alice and Bob share the same instance of analytics_pipeline in staging
1. Policy & rules
    - Different people have different access rights within Aptomi object model
    - Aptomi can allocate instances with different parameters in different clusters (e.g. based on consumer identity)

# Instructions

1. First of all, bootstrap Aptomi on behalf on Sam (domain admin) to import k8s clusters and rules. ACL rules are defined in such
a way that Sam is a domain admin, John/Frank are namespace admins, and Alice/Bob/Carol are service consumers.

    To feed cluster information into Aptomi, copy clusters template file and enter kubecontext configuration into it:
    ```
    cp examples/03-twitter-analytics/policy/Sam/clusters.{yaml.template,yaml}
    vi examples/03-twitter-analytics/policy/Sam/clusters.yaml
    ```

    If you are using the provided `./tools/demo-gke.sh` script, get clusters configuration by running the following command and paste output into clusters.yaml.
    Make sure to (1) copy config for each cluster separately, (2) put us-east -> us-east and us-west -> us-west correctly:
    ```
    ./tools/demo-gke.sh kubeconfig
    ```

    Upload the list of clusters and ACL rules into Aptomi using CLI:
    ```
    aptomictl policy apply --username Sam -f examples/03-twitter-analytics/policy/Sam
    ```
1. Import analytics_pipeline service definition on behalf of Frank
    ```
    aptomictl policy apply --username Frank -f examples/03-twitter-analytics/policy/Frank
    ```
1. Import twitter_stats service definition on behalf of John
    ```
    aptomictl policy apply --username John -f examples/03-twitter-analytics/policy/John
    ```
1. At this point all service definition have been published to Aptomi, but nothing has been instantiated yet. You can see
that in Aptomi UI under "Policy Browser"

1. Now, we need to provide user secrets, so twitter_stats component can pull data over Twitter Streaming API. Create 3
applications in [Twitter Application Management Console](https://apps.twitter.com)
    ![Twitter App Create](twitter-app-create.png)
    
    Generate keys and access tokens for them:
    ![Twitter Create Tokens](twitter-create-tokens.png)
    
    Once done, copy secrets.yaml and enter the created keys/tokens into it:
   ```
   cp examples/03-twitter-analytics/_external/secrets/secrets.yaml.template /etc/aptomi/secrets.yaml
   vi /etc/aptomi/secrets.yaml
   ```

1. Now let's have consumers declare 'dependencies' on the services defined by John and Frank. John requests an instance
    ```
    aptomictl policy apply --wait --username John -f examples/03-twitter-analytics/policy/john-prod-ts.yaml
    ```

    Aptomi allocates dedicated production instance in cluster `cluster-us-east` according to the rule `analytics_prod_goes_to_us_east` defined in [rules.yaml](policy/Sam/rules.yaml).
    It instantiates all required services according to the service graph and deploys them to the corresponding k8s cluster.
    You can navigate to "Policy Browser", select "Desired State" and inspect how allocation was performed and which services got instantiated.
    Also, on the "Instances" tab, you can retrieve all endpoints for the deployed service.

    To check deployment progress, you can run the following command and wait for "1/1" status for pods:
    ```
    watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods
    ```


1. Alice and Bob request instances
    ```
    aptomictl policy apply --wait --username Alice -f examples/03-twitter-analytics/policy/alice-stage-ts.yaml
    aptomictl policy apply --wait --username Bob -f examples/03-twitter-analytics/policy/bob-stage-ts.yaml
    ```
    We are assuming that Alice is a developer and she wants to test a different version of visualization code for twitter-stats.
    Bob is just a service consumer that wants to instantiate the same service, but look at the tweets from Mexico.

    Aptomi allocates those service instances in cluster `cluster-us-west` according to the rule `analytics_stage_goes_to_us_west` defined in [rules.yaml](policy/Sam/rules.yaml).
    Moreover, those services are sharing the same analytics pipeline, as you can see in UI under "Policy Browser" in "Desired State".
    Similarly, on the "Instances" tab, you can retrieve all endpoints for the deployed services.

    To check deployment progress, you can run the following command and wait for "1/1" status for pods:
    ```
    watch -n1 -d -- kubectl --context cluster-us-west -n demo get pods
    ```

1. If everything got deployed successfully, then you should be able to see 
   - running *production* instance of twitter stats in `cluster-us-east` (managed by John)
   - running *staging* instance of twitter stats in `cluster-us-west` (Alice's version with new look & feel)
   - running *staging* instance of twitter stats in `cluster-us-west` (Bob's version with tweets from Mexico)

   If you open tweeviz HTTP endpoints, you will be able to see tweets coming in real-time over Twitter Streaming API, getting processed through analytics-pipeline, and displayed on the web.

1. Now, let's demonstrate an update of a running instance of in production.
    Alice can never deploy to production cluster according to the rules defined in Aptomi, so Alice tells Aptomi to tear down her staging instance and asks John to update production instance.

    Alice removes her twitter-stats instance which runs in staging:
    ```
    aptomictl policy delete --wait --username Alice -f examples/03-twitter-analytics/policy/alice-stage-ts.yaml
    ```
    John changes label for twitter-stats instance which runs in production:
    ```
    sed -e 's/demo11/demo12/g' examples/03-twitter-analytics/policy/john-prod-ts.yaml > examples/03-twitter-analytics/policy/john-prod-ts-changed.yaml
    aptomictl policy apply --wait --username John -f examples/03-twitter-analytics/policy/john-prod-ts-changed.yaml
    ```

    After that, if you reload tweeviz HTTP endpoints in the browser, you will see that:
    - Alice's stating version is no longer available
    - John's production version runs a new look & feel that Alice was testing

1. Carol belongs to 'mobile-dev' team, so she cannot instantiate any services according to the rule `reject_dependency_for_mobile_dev_users` defined in [rules.yaml](policy/Sam/rules.yaml).
    ```
    aptomictl policy apply --wait --username Carol -f examples/03-twitter-analytics/policy/carol-stage-ts.yaml
    ```
