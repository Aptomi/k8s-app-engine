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

3. Initialize demo repository on GitHub and start auto-apply
    - ./tools/demo-init.sh
    - ./tools/demo-push.sh
    - ./tools/demo-watch-apply.sh

### Demo scenario

1. Explain initial state
    - No services have been defined
    - Aptomi only knows about 2 kubernetes clusters (us-east and us-west) and 5 users

2. Frank is an Service Ops guy. He comes and defines analytics-pipeline service in Aptomi
    - He defines all components of this service and specifies where the containers are
    - He defines contexts for his service
    - Context is our secret sauce. It describes who the service is for and how resources are allocated/shared
    - In this case Frank offers analytics-pipeline in 2 contexts
      - for analytics_ops_team (who control production instance)
          - for them the service will run in cluster-us-east
      - for development (who will share instance of this service)
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
      - for analytics_ops_team (who control production instance)
          - for them the service will run in cluster-us-east. same as Frank's
      - for development (who will
          - for them the service will run in cluster-us-west
          - but every developer

    - enabled = true in Frank/analytics_pipeline/service.analytics_pipeline.yaml
    - Show UI - audit log
    - Show UI - delta picture
    - New service definition has been published to aptomi

## Suntrust demo / story:

2. Show policy
   - Show k8s clusters
      - `kubectl config get-contexts`
      - `kubectl config view`

3. Deploy AP + TS for user Alice (ID=1)
   - Explain what will get matched (Alice doesn't specify, policy controls that)
     - low-Alice (priority < 200), team-platform-services (priority < 200)
   - Run aptomi
     - `./aptomi policy apply --noop`
     - `./aptomi policy apply --noop --show`
     - `./aptomi policy apply`
   - Run aptomi again
     - `./aptomi policy apply` - to ensure there are no more changes to apply
   - While it's loading, we can show tracing
     - `./aptomi policy apply --noop --trace`
   - Show kubectl output
     - `kubectl --context cluster-us-west -n demo get pods`
     - `watch -n1 -d -- kubectl --context cluster-us-rwest -n demo get pods`
   - Show endpoints
     - `./aptomi endpoint show`
   - Open Tweeviz UI
     - Shows SF, NY, Boston tweets

4. Deploy AP + TS for user Bob (ID=2). Alice (ID=1) deploys new staging version of TS (Canary testing/updates)
  - Alice - with "demo-v42" and "stage" as tags
  - Explain what will get matched
     - low-Bob (priority < 200), team-platform-services (priority < 200)
     - meaning, dedicated TS and shared AP
   - Run aptomi
     - `./aptomi policy apply --noop`
     - `./aptomi policy apply --noop --show` (explain share and reuse of services)
     - `./aptomi policy apply`
   - Show endpoints
     - `./aptomi endpoint show`
   - Open Tweeviz UI
     - Show stage Alice (different UI)
     - Show prod Bob (Japan tweets)

5. Alice (ID=1) propagates staging version to production (making a change)
   - Staging TS gets deleted
   - Production TS gets updated
   - Make sure to use `./aptomi policy apply --noop --show --verbose` to see deletions and updates
   - Refresh in browser
     - Stage instance disappears
     - Prod instance changes look and feel to demo-v42

6. Change # of top tweets
   - Default -> 3 and redeploy

7. Alice (ID=1) gets marked as "compromised"
   - Loses access to her "prod"
   - Right now all objects get deleted, but this behavior will be customizable

8. Low-priority user Carol (ID=3)
   - Gets nothing due to priority < 50

9. Deploy dedicated DAP for Carol (ID=3) in its own k8s cluster
   - Priority = 200
   - Show kubectl output
     - `kubectl --context cluster-us-east -n demo get pods`
     - `watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods`
