# oneshot

This is a very tiny, unix style program to solve one need: automatic task delivery.

## Why


During my career I was in charge of the recruitment process. Part of it was the algorithmic challenge. Applicant is tasked to download the assessment  (containing the task description and example input and output) and solve it in given time. When the applicant download the assessment the event is logged. The recruiter has an easy way to check when that happen. This way the recruitment is free to schedule and record the task delivery and time tracking.

Other solution is to use one of the paid services from the market.

## Installation

    go get github.com/robert-zaremba/oneshot

Make sure that your `$GOBIN` is in your `$PATH`.

## Usage

`onshot` is a webservice, accessible through HTTP.

Running:

    ./oneshot  <port_to_listen> <admin_password>

You can use [curl](https://curl.haxx.se/) to communicate with the webservice.

To create a new job:

     curl -i -H "X-Auth-Token: <admin_password>" "http://your.domain:<port>/oneShot/newjob/?name=<new_job_name>" -F file=@./some_file.tgz

To assign a task:

     curl -i -H "X-Auth-Token: <admin_password>" "http://your.domain:<port>/oneShot/assign/?job=<job_name>&user=<username>"

This will return the unique task ID.

To download the task:

    your_favourite_browser "http://your.domain:<port>/oneShot/gettask/?task=<task_id>"

Be aware - once the task is downloaded it can't be downloaded any more. You will need to create a new assignment. If you will use the link the second time (eg after receiving the task solution) the service will return the UTC time when the task was downloaded.


## License:

Apache License v2.0

© Robert Zaremba
