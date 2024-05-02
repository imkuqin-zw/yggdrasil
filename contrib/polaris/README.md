# yggdrasil with Polaris

English | [简体中文](./README-zh.md)

## Introduction

yggdrasil-polaris provides a series of components based on yggdrasil framework, developers can use yggdrasil-polaris to build distributed micro service applications.

## Key Features

* **Service Registration and Heartbeat**: To register the service and send heartbeat periodly.
* **Service Routing and LoadBalancing**: Implement yggdrasil resover and balancer, providing semantic rule routing and loadbalacing cases.
* **Fault node circuitbreaking**: Kick of the unhealthy nodes when loadbalacing, base on the service invoke successful rate.
* **RateLimiting**: Implement yggdrasil interceptor, providing request ratelimit check to ensure the reliability for server.