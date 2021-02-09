## Environment based configuration

* `COUCHDB_URL` default: `"http://couchdb:5984/"`
    * URL to connect to couchdb, http or https.
* `COUCHDB_USER` default: `"admin"`
    * User used to authenticate with couchdb.
* `COUCHDB_PASSWORD` default: `"secret"`
    * Password user to authenticate with couchdb.
* `COUCHDB_DB` default: `"test"`
    * Default database to connect to.
* `COUCHDB_DB_AUTO_CREATE` default: `"true"`
    * If enabled will create the couchdb database if possible.
* `COUCHDB_HEALTH_CHECK_KEY` default: `"$health_check"`
    * Name of the object that is used for the health check operation.
* `COUCHDB_HEALTH_CHECK_RESULT_TTL` default: `10s`
    * Amount of time to cache the last health check result.
