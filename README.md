GitOps Controller
====

This controller is helper controller that synchronizes the ECR repository with the k8s manifest repository.

## Description

The purpose of this controller is to be able to synchronize the Docker registry with the manifest's Git repository, just like Weave Flux on Argo-CD.  
In Weave Flux, to synchronize with multiple Git repositories, you need to start flux daemons as many as the number of Git repositories, but this controller does not need to do so.  

https://github.com/fluxcd/flux/issues/1164

## Usage

ToDo

## Licence

MIT

## Author

[Kazylla](https://github.com/Kazylla)
