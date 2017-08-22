## Demo slides:
    - https://docs.google.com/a/renski.com/presentation/d/10X8oEjBqLPSfxf5KHhg8n_tBtYa7bAERHkZZiynjwyc/edit?usp=sharing

1. Aptomi is built around a concept of service

2. Demo setup. As an example John set up rules for twitter-stats in a way that
   - single instance in prod
   - one instance for every developer in stage

   On GKE (two regions). Will be instantiated and deployed for real.

3. People with different roles interact with Aptomi
   - publish & define rules
   - consume
   - governance & set up rules of the land

   Aptomi processes this information, evaluates policy, makes decisions, instantiates services and configures components in the underlying cloud if/as needed
   Aptomi provides visibility for Ops into what's going on

## Demo steps

### Preparation

1. Switch to the Aptomi GCloud account and project

For Sergey:
```shell
gcloud config set account sergei@apptomi.com
gcloud config set project symmetric-fin-171202
```

For Roman:
```shell
gcloud config set account roman@apptomi.com
gcloud config set project bright-torus-169502
```

2. Create k8s clusters
    - `./tools/demo-gke.sh up`
    - `./tools/demo-gke.sh cleanup`

    - For faster cleanup, use:
      - Fast cleanup in west: `kubectl --context cluster-us-west -n demo delete ns demo`
      - Fast cleanup in east: `kubectl --context cluster-us-east -n demo delete ns demo`
      - Delete helm objects: `./tools/demo-gke.sh cleanup`

3. Terminal Tab #1: initialize local DB, initialize demo repository on GitHub and start watcher/auto-applier
    - `cd tools/demo_0_0_3`
    - `./demo-init.sh`

4. Terminal Tab #2: run UI server
    - `cd tools/demo_0_0_3`
    - `./demo-start-ui.sh`

5. Terminal Tab #3: run k8s monitoring (split tab horizontally)
    - `watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods`
    - `watch -n1 -d -- kubectl --context cluster-us-west -n demo get pods`

6. Ideally, run the initial deployment once to cache all container images

### Demo scenario

1. Explain initial state
    - No services have been defined
    - Aptomi only knows about 2 kubernetes clusters (us-east and us-west) and 5 users

2. Frank is an Service Ops guy
    - He just recently started using Aptomi. He quickly described services that he owns
        - kafka, spark, zookeeper, hdfs
        - analytics-pipeline, which is a service we will focus on. It's a service that embeds the other 4 under one umbrella
    - He specified "service definitions"
        - Where docker container images are
    - He defined "contexts for his service"
        - Context is our secret sauce
        - It describes who the service is for and how resources are allocated/shared
    - In this case Frank offers analytics-pipeline in 2 contexts
      - for it ops (who control production instance)
          - for them the service will run in cluster-us-east
      - for developers (who will share instance of this service)
          - for them the service will run in cluster-us-west
    - enabled = true in Frank/analytics_pipeline/service.analytics_pipeline.yaml
    - Show UI - audit log
    - Show UI - delta picture
    - New service definition has been published to aptomi
        - analytics pipeline, all its dependencies, and contexts

3. John is another Service Ops guy. He comes and defines twitter-stats service in Aptomi
    - John's service relies on Frank's service analytics-pipeline
      - and it contains additional components for reading/processing/visualizing messages from twitter stream
    - John offers twitter-stats in 2 contexts as well
      - for it ops (who control production instance)
          - for them the service will run in cluster-us-east. same as Frank's
      - for developers
          - for them the service will run in cluster-us-west
          - but when they request an instance, every developer will get its own instance of twitter-stats
          - as opposed to sharing it
    - enabled = true in John/twitter_stats/service.twitter_stats.yaml
    - Show UI - audit log
    - Show UI - delta picture
    - New service definition has been published to aptomi

4. At this point Aptomi just has service definitions and nothing has been instantiated yet
    - Now let's have some consumers declare dependencies on the services defined by John and Frank

[PAUSE / SEE IF SOMEONE HAS ANY QUESTIONS]

5. Declaring dependencies (Prod)
    - Production
        - let Ops team to instantiate twitter-stats service in production
    - enabled = true in John/dependencies.john-ts.yaml
    - Show UI - audit log
    - Show UI - delta picture
    - Show containers on k8s

6. Declaring dependencies (Dev)
    - Developers
        - let Alice to instantiate twitter-stats service
            - Alice tests a new web app code for visualizing data from twitter
            - enabled = true in Alice/dependencies.alice-stage-ts.yaml
        - let Bob instantiate twitter-stats service
            - enabled = true in Bob/dependencies.bob-stage-ts.yaml
        - Note
            - neither Alice or Bob define the specifics (where this service should run, whether they want to share an instance or not, etc)
            - Alice & Bob just declare their intents (saying "I want to consume this service"). Aptomi makes all allocation decisions for them
            - Implementation details are abstracted away from the consumer. The way how service is allocated is controlled by a service owner via contexts.
    - Show UI - audit log
    - Show UI - delta picture
    - Show containers on k8s

7. Now, what happened exactly
    - Show global view
    - From the standpoint of consumer
        - Alice: started to consume the service
        - All dependencies got magically allocated for them
    - From the standpoint of service owner
        - Frank: seeing that his analytics_pipeline service got allocated 2 times (per the rules he defined)
        - Show Home Page
        - Show Policy Explorer (e.g. service view from the standpoint of zookeeper)
    - John is seeing that his twitter_stats service got allocated to 3 consumers (per the rules he defined)
        - Show Home Page
        - Show Policy Explorer (twitter stats that he OWNS)
        - Show Policy Explorer (twitter stats that he RUNS)
    - Show that everything got deployed and working
        - John opens production endpoints for twitter stats
        - Alice opens dev endpoints for twitter stats (different visualization)
        - Bob opens dev endpoints for twitter stats (standard visualization, but Mexico)
        - (!) Leave production endpoint with twitter stats opened in browser

[PAUSE / SEE IF SOMEONE HAS ANY QUESTIONS]

6. Now that we know the concepts, let's show some of the more advanced functionality
    - updating a service
    - not allowing certain users/teams to consume a service
    - blocking access to a service

7. Updating production instance of twitter stats
    - Now let's say Alice is happy with her change and this change needs to be rolled out to production
    - There is no way Alice can deploy to production directly by herself
        - Aptomi will actually never allow that, because of the global rule Dev -> only Dev cluster
        - Show rule
        - So Alice has to ask John, who owns production instance
    - So Alice gets rid of her instance
        - enabled = false in Alice/dependencies.alice-stage-ts.yaml
    - John promotes new version of visualization app to production
        - tsvisimage: demo-v62 in John/dependencies.john-ts.yaml

8. Show rejected access. Carol tries to get her service twitter_stats
   - enabled = true in Carol/dependencies.carol-stage-ts.yaml
   - Nothing will get allocated because of the global rule
   - Show Carol's home page
   - Show Policy Explorer
   - Show Rule Log

9. Bob gets marked as "deactivated"
   - deactivated = true in LDAP
   - make a dummy check-in into Github (to generate a commit)
   - Loses access to his instance (via Istio)

10. Show Rules on home page from Sam's point of view
   - What rules get applied where

11. That concludes the demo
   - This was 100% live, deployed from scratch on Google Cloud
