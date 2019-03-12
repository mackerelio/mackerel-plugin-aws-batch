# mackerel-plugin-aws-batch [![Build Status](https://travis-ci.org/mackerelio/mackerel-plugin-aws-batch.svg?branch=master)](https://travis-ci.org/mackerelio/mackerel-plugin-aws-batch)

## Install

```sh
% mkr plugin install mackerelio/mackerel-plugin-aws-batch
```

## Setting

```
[plugin.metrics.aws-batch]
command = "/path/to/mackerel-plugin-aws-batch -access-key-id=XXXXX -secret-access-key=XXXXX -job-queue=MyJobQueue1 -job-queue=MyJobQueue2 -region=ap-northeast-1"
```
