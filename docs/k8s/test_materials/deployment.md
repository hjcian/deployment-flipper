

POD is deploying...
```json
{
  "type": "Available",
  "status": "False",
  "lastUpdateTime": "2022-07-23T08:32:29Z",
  "lastTransitionTime": "2022-07-23T08:32:29Z",
  "reason": "MinimumReplicasUnavailable",
  "message": "Deployment does not have minimum availability."
}
{
  "type": "Progressing",
  "status": "True",
  "lastUpdateTime": "2022-07-23T08:32:29Z",
  "lastTransitionTime": "2022-07-23T08:32:29Z",
  "reason": "ReplicaSetUpdated",
  "message": "ReplicaSet \"deployment-flipper-demo-98875cf84\" is progressing."
}
```


POD is running.
```json
{
  "type": "Available",
  "status": "True",
  "lastUpdateTime": "2022-07-23T08:32:33Z",
  "lastTransitionTime": "2022-07-23T08:32:33Z",
  "reason": "MinimumReplicasAvailable",
  "message": "Deployment has minimum availability."
}
{
  "type": "Progressing",
  "status": "True",
  "lastUpdateTime": "2022-07-23T08:32:33Z",
  "lastTransitionTime": "2022-07-23T08:32:29Z",
  "reason": "NewReplicaSetAvailable",
  "message": "ReplicaSet \"deployment-flipper-demo-98875cf84\" has successfully progressed."
}
```