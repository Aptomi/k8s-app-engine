## Demo slides:
    - https://docs.google.com/a/renski.com/presentation/d/10X8oEjBqLPSfxf5KHhg8n_tBtYa7bAERHkZZiynjwyc/edit?usp=sharing

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

3. Initialize local DB, initialize demo repository on GitHub and start watcher/auto-applier
    - `./tools/demo-init.sh`

4. Run server
    - `./tools/dev-watch-server.sh`

5. Run k8s monitoring in background in two separate tabs
    - `watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods`
    - `watch -n1 -d -- kubectl --context cluster-us-west -n demo get pods`

6. Ideally, run the initial deployment once to cache all container images

### Demo scenario

1. Explain initial state
    - No services have been defined
    - Aptomi only knows about 2 kubernetes clusters (us-east and us-west) and 5 users

2. Frank is an Service Ops guy. He comes and defines analytics-pipeline service in Aptomi
    - He defines all components of this service and specifies where the containers are
    - He defines contexts for his service
    - Context is our secret sauce. It describes who the service is for and how resources are allocated/shared
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

5. Declaring dependencies (Prod)
    - Production
        - let Ops team to instantiate twitter-stats service in production
    - enabled = true in John/dependencies.john-ts.yaml
    - Show UI - audit log
    - Show UI - delta picture
HERE GCLOUD TOKEN SHIT BROKE AGAIN IN WATCHER !!!! !!!! !!!!
    - Show containers on k8s
        - watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods

6. Declaring dependencies (Dev)
    - Developers
        - let Alice to instantiate twitter-stats service
            - Alice tests a new web app code for visualizing data from twitter
            - enabled = true in Alice/dependencies.alice-stage-ts.yaml
        - let Bob instantiate twitter-stats service
            - enabled = true in Bob/dependencies.bob-stage-ts.yaml
    - Show UI - audit log
    - Show UI - delta picture
    - Show containers on k8s
        - watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods

7. Now, what happened exactly
    - Frank is seeing that his analytics_pipeline service got allocated 2 times (per the rules he defined)
        - Log in as John
        - Show Home Page
        - Show Policy Explorer (e.g. service view from the standpoint of zookeeper)
    - John is seeing that his twitter_stats service got allocated to 3 consumers (per the rules he defined)
        - Log in as John
        - Show Home Page
        - Show Policy Explorer (twitter stats that he OWNS)
        - Show Policy Explorer (twitter stats that he RUNS)
    - Show that everything got deployed and working
        - John opens production endpoints for twitter stats
        - Alice opens dev endpoints for twitter stats (different visualization)
        - Bob opens dev endpoints for twitter stats (standard visualization, but Mexico)

6. Updating production instance of twitter stats
    - Now let's say Alice is happy with her change and this change needs to be rolled out to production
    - There is no way Alice can deploy to production directly by herself
        - Aptomi will actually never allow that, because of the global rule Dev -> only Dev cluster
        - So Alice has to ask John, who owns production instance
    - So Alice gets rid of her instance
        - enabled = false in Alice/dependencies.alice-stage-ts.yaml
    - John promotes new version of visualization app to production
        - tsvisimage: demo-v62 in John/dependencies.john-ts.yaml

7. Show rejected access. Carol tries to get her service twitter_stats
   - enabled = true in Carol/dependencies.carol-stage-ts.yaml
   - Nothing will get allocated because of the global rule
   - Show Carol's home page
   - Show Policy Explorer
   - Show Rule Log

8. Bob gets marked as "deactivated"
   - deactivated = true in _external/users/users.dev.yaml
   - Loses access to his instance (via Istio)

9. Show Rules on home page from Sam's point of view
   - What rules get applied where
