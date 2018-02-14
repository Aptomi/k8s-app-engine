# Description

In this example there are 2 key services:
* `analytics_pipeline` - it's data analytics service which consists of Kafka, Spark, HDFS, and Zookeeper
* `twitter_stats` -  it's a Twitter data processing service, which gets messages from Twitter in real time, calculates top hash tag statistics using an external `analytics_pipeline` service
   and displays results on a web page

These services are offered in two different contexts/flavors:
* *stage*: intended for development and testing. A single instance of `analytics_pipeline` gets shared between all consumers. `twitter_stats` uses a mock implementation, which generates a stream of random Twitter messages.
* *production*: intended for production use. `analytics_pipeline` with better availability and performance. `twitter_stats` receives messages from Twitter in real time

![Diagram](diagram.png)

These services have different owners, who can fully control how their services get offered and shared:

User  | Role |
------|-------
Sam   | Domain admin for Aptomi
Frank | Owner of `analytics_pipeline`
John  | Owner of `twitter_stats`
Alice, Bob, Carol | Users and code contributors to `twitter_stats`

This example illustrates a few important things that Aptomi does:

1. **Run multi-component apps across several environments**
    * Define an application and its contexts/flavors
    * Components get lazily instantiated based on the service/component graph
    * Update components of an application after you deploy without affecting dependent services
1. **Service ownership & reuse**
    - Owners define how their services get offered & shared with others
    - Consumers can be directed to use the same service instance
1. **Control**
    - Define who can have access to which services

# Instructions

1. Upload user roles and rules into Aptomi using CLI, then import `analytics_pipeline` and `twitter_stats` services:
    ```
    aptomictl login -u sam -p sam
    aptomictl policy apply --wait -f ~/.aptomi/examples/twitter-analytics/policy/Sam
    aptomictl login -u frank -p frank
    aptomictl policy apply --wait -f ~/.aptomi/examples/twitter-analytics/policy/Frank
    aptomictl login -u john -p john
    aptomictl policy apply --wait -f ~/.aptomi/examples/twitter-analytics/policy/John
    ```

1. At this point all service definition have been published to Aptomi, but nothing has been instantiated yet. You can see
that in Aptomi UI under [Policy Browser](http://localhost:27866/#/policy/browse)

1. Start two staging instances of `twitter-stats`, wait until they are up:
    ```
    aptomictl login -u alice -p alice
    aptomictl policy apply --wait -f ~/.aptomi/examples/twitter-analytics/policy/alice-stage-ts.yaml
    aptomictl login -u bob -p bob
    aptomictl policy apply --wait -f ~/.aptomi/examples/twitter-analytics/policy/bob-stage-ts.yaml
    ```

    To check deployment progress, run `kubectl get pods` and wait for "1/1" status for all created pods:
    ```
    watch -n1 -d -- kubectl -n west get pods
    ```

    Note that first-time deployment will download all Helm charts & container images (1GB+) and cache them in your k8s cluster(s). This can take a while, depending on your internet connection (pods will be in `ContainerCreating` status)
      * k8s on GKE will give you the best experience (5-10 min)
      * running locally on Minikube or Docker For Mac will likely be slower (up to 15-20 min)

    All subsequent runs will be much faster, as charts and images will be already there.

1. Start production instance of `twitter-stats`:

    ```
    aptomictl login -u john -p john
    aptomictl policy apply --wait -f ~/.aptomi/examples/twitter-analytics/policy/john-prod-ts.yaml
    ```

    To check deployment progress, run `kubectl get pods`:
    ```
    watch -n1 -d -- kubectl -n east get pods
    ```

    **Important note**: you will need to take an extra step in order to get *production* environment fully up and running. Right now `tweepub` and `tweeviz` health checks will be showing "0/1" - this
    is expected, as they can't use Twitter Streaming API because tokens have not been configured. Once you set up tokens in Twitter and pass them
    to Aptomi, those services will start successfully. See the next section on how to configure Twitter keys and pass them to Aptomi.

1. At this point you can see that:
    * Aptomi applied defined [rules](policy/Sam/rules.yaml) and allocated *stage* in `cluster-us-west` and *production* in `cluster-us-east`

    * Aptomi dictated that `analytics pipeline` will be shared in staging by both Alice and Bob. See [Policy Browser](http://localhost:27866/#/policy/browse) -> Desired State

    * Service endpoints are available under [Instances](http://localhost:27866/#/policy/dependencies) in Aptomi UI
        * `tweeviz` available in *stage* over HTTP with fake Twitter data source
        * `tweeviz` available in *prod* over HTTP with real-time Twitter data (assuming you have followed the step to configure Twitter tokens)

# Advanced: Enabling Streaming Data from Twitter

1. If you want a truly fully functional demo, you can actually inject Twitter App Tokens into *production* instance of `twitter_stats`, so it can start pulling data from Twitter Streaming API. Create an
application in [Twitter Application Management Console](https://apps.twitter.com)
    ![Twitter App Create](twitter-app-create.png)

    Generate keys and access tokens for it:
    ![Twitter Create Tokens](twitter-create-tokens.png)

    Once done, upload the updated *production* dependency with Twitter tokens into Aptomi:
    ```
    cp ~/.aptomi/examples/twitter-analytics/policy/john-prod-{ts.yaml,ts-changed.yaml}
    vi ~/.aptomi/examples/twitter-analytics/policy/john-prod-ts-changed.yaml
    aptomictl login -u john -p john
    aptomictl policy apply --wait -f ~/.aptomi/examples/twitter-analytics/policy/john-prod-ts-changed.yaml
    ```
Now if you open HTTP endpoint for *production* instance of `tweeviz`, you will be able to see hash tag statistics calculated over real-time messages retrieved over Twitter Streaming API.

