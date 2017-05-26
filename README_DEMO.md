## Demo scenario:

1. Explain use case (show slides)
   - "Data Analytics Pipeline" service (kafka, spark, hdfs, zookeeper) for data processing
   - "Twitter Stats" service (twitter real-time messages -> DAP -> stats on web)
   - Explain our model (what gets matched and how)
   - Explain what we are going to show

2. Show policy
   - Run with empty dependencies. Show service graph
   - Show service definition files for DAP
   - Show service definition files for TS
   - Show contexts/allocations
     - Explain what contexts will be matched for which users
   TODO: show k8s cluster definition files and multi-kubernetes cluster support

3. Deploy DAP + TS for user Alice (ID=1)
   - Single DAP, single TS
     - policy apply --noop
     - policy apply --noop --verbose
     - picture of services graph
     - show visualization on web (get endpoints via aptomi)
     - policy apply

4. Deploy DAP + TS for user Bob (ID=2)
   - Share the same DAP, dedicated TS
     - policy apply --noop
     - policy apply --noop --verbose
     - picture of services graph
     - show visualization on web (get endpoints via aptomi)
     - policy apply

5. Low-priority user Carol (ID=3)
   - Gets nothing due to low priority

6. Alice (ID=1) deploys new staging version of TS in parallel
   - Show TS
     - tweepub: different region
     - tweeviz: different HTML to visualize data

7. Alice (ID=1) propagates staging version to production
   - Staging TS gets deleted
   - Production TS gets updated

8. Alice (ID=1) gets marked as "untrusted"
   - Loses access to her "prod"
