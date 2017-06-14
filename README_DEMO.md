## Demo scenario:

0. Cleanup env

```shell
tools/demo-gke.sh cleanup
tools/demo-gke.sh status
rm -rf db/*
make clean build test install
```

1. Show slides
  - https://docs.google.com/presentation/d/17OeSqPzOC8ng1Roe3hkGVWvLJxa-o-VI3jp1GLXcCt0/edit?usp=sharing

2. Init demo. Frank & John run prod
  - `./tools/demo-init.sh` to load all objects into Aptomi
  - `aptomi policy apply --show` to deploy initial state
  - Initial state is
    - Prod is in cluster-us-east
      - Frank (Ops) shares prod instance of twitter_stats with the org
      - John (Ops) shares prod instance of analytics_pipeline with org
  - Show that it got deployed on k8s
    - `watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods`
  - Run aptomi again
    - `aptomi policy apply --noop` - to ensure there are no more changes to apply
  - Show endpoints
    - `aptomi endpoint show`
  - Open Tweeviz UI
    - Shows SF, NY, Boston tweets

3. Alice & Bob deploy stage
  - Alice deploys new staging version of TS viz (Canary testing/updates)
  - Bob deploys new staging version of TS (Mexico tweets)
  - Run aptomi
    - `aptomi policy add dependencies demo/dependencies/dependencies.alice-stage-ts.yaml`
    - `aptomi policy add dependencies demo/dependencies/dependencies.bob-stage-ts.yaml`
    - `aptomi policy apply --noop --show` (explain share and reuse of services)
    - `aptomi policy apply`
  - Explain
    - Stage is in cluster-us-west
    - Alice & Bob have their own twitter apps
    - Alice & Bob share analytics pipeline
  - Show that it got deployed on k8s
    - `watch -n1 -d -- kubectl --context cluster-us-west -n demo get pods`
  - Run aptomi again
    - `aptomi policy apply --noop` - to ensure there are no more changes to apply
  - Show endpoints
    - `aptomi endpoint show`
  - Open Tweeviz UI
    - Show stage Alice (different UI)
    - Show prod Bob (Mexico tweets)

4. Alice deletes her staging instance and asks Frank to propagate her new VS to production
  - Run aptomi
    - `aptomi policy delete dependencies demo/dependencies/dependencies.alice-stage-ts.yaml`
    - `vim demo/dependencies/dependencies.frank-prod-ts.yaml` and change demo-v51 -> demo-v52
    - `aptomi policy add dependencies demo/dependencies/dependencies.frank-prod-ts.yaml`
    - `aptomi policy apply --noop` - show that there are some deleted services and some to update
    - `aptomi policy apply`
   - Refresh in browser
     - Stage instance disappears
     - Prod instance changes look and feel to demo-v52

5. Carol deploys her staging instance of TS
  - Run aptomi
    - `aptomi policy add dependencies demo/dependencies/dependencies.carol-stage-ts.yaml`
    - `aptomi policy apply --show` - there are no changes because Carol is compromised
  - Show ```demo/rules/rules.compromised-users.yaml``` rule and explain that global rule doesn't allow Carol to deploy
    because she's compromised
  - Run aptomi again
    - `vim demo/users/users.dev.yaml` and remove line `compromised: true` from Carol's labels
    - `aptomi policy add users demo/users/users.dev.yaml` - updated users
    - `aptomi policy apply --show` - show that now services for Carol will be deployed
    - `aptomi endpoint show`
  - Show that it got deployed on k8s
    - `watch -n1 -d -- kubectl --context cluster-us-west -n demo get pods`
  - Open Tweeviz UI
    - Show stage Carol (Brazil tweets)
    
6. Cluster-us-west gets compromised and all exposed service will be blocked by Istio
  - Run aptomi
    - `vim demo/clusters/cluster.cluster-us-west.yaml` - uncomment `compromised: true` label
    - `aptomi policy add cluster demo/clusters/cluster.cluster-us-west.yaml` - update cluster
  - Show ```demo/rules/rules.compromised-clusters.yaml``` rule and explain that global rule will block ingress access
    to all services on compromised cluster
    - `aptomi policy apply` - apply changes
    
7. Demonstrate that dev & prod are separate (e.g. Dev can't deploy to Prod)
  - Show `demo/services/twitter_stats/context.prod.twitter_stats.yaml` - it'll accept only operators
  - Show `demo/rules/rules.dev-users.yaml` - global rule restricts dev org users use of prod clusters 
