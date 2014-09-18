sharded_counter
==============

Go Google App Engine scalable sharding counters.

Copied from [appengine articles](https://developers.google.com/appengine/articles/sharding_counters)

> ...it is important to note that you can only expect to update any single entity or entity group about five times a second

> If you had a single entity that was the counter and the update rate was too fast, then you would have contention as the serialized writes would stack up and start to timeout. The way to solve this problem is a little counter-intuitive if you are coming from a relational database; the solution relies on the fact that reads from the App Engine datastore are extremely fast and cheap. The way to reduce the contention is to build a sharded counter – break the counter up into N different counters.

Sharding counter is an alternative solution for count the number of votes in a poll, the number of comments, or even the number of visitors to your site.
