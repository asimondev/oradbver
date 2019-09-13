# oradbver

Use this Go program to get the Oracle Database version in JSON format. Additionally the 
parameter **-ping** pings the specified database every 1 second.

 ## Working with oradbver
 
 ### Before starting oradver
 
 oradbver uses the Oracle OCI client library. That means, that either Oracle database software or 
 the client software must be installed. The environment parameters must be set as well. 
 As a test the SQL*Plus must run without any problems.
 
 ### Parameters
 
 - **-u** database user. You can omit this parameter to use the "internal" connect.
 - **-p** database user password. The password will be asked, if you have specified the 
 database user.
 - **-d** database. This could be easy connect string or a TNS alias.
 - **-r** system privilege. At the moment only "sysdba" is supported.
 - **-c** JSON configuration file. You can specify all corresponding parameters in this file.
 - **-ping** Start pinging the database every second. You can stop this by pressing Return key.
 - **-ping-once** Ping the database once and exit with usual UNIX (0 or 1) return code.
  
 Here is an example of the JSON configuration file:
 
    {
       "user": "sys",
       "password": "oracle",
       "database": "orcl",
       "role": "sysdba"
    }


## Examples

### Running oradbver without parameters using the internal connect.

    > ./oradbver
    {"Release":"11.2.0.4","Version":11,"RAC":false,"CDB":false,"Role":"PRIMARY"}

### Running oradbver with parameters.

    > ./oradbver -u sys -r sysdba -d a01.world
    Enter password: 
    {"Release":"11.2.0.4","Version":11,"RAC":false,"CDB":false,"Role":"PRIMARY"}

### Running oradbver with JSON configuration file.

#### Using local connect without listener.
    > cat ~/tmp/db.json
    {
      "user": "andrej",
      "password": "andrej"
    }
    
    > ./oradbver -c ~/tmp/db.json
    {"Release":"11.2.0.4","Version":11,"RAC":false,"CDB":false,"Role":"PRIMARY"}

#### Using remote connect with database name as TNS connection string.

The database value may not contain any whitespace characters.

    {
      "user": "sys",
      "password": "oracle",
      "role": "sysdba",
      "database": "(description=(address=(protocol=tcp)(host=avmol7db1)(port=1521))(connect_data=(service_name=a01.de.oracle.com)))"
    }

### Running oradbver with JSON configuration file and **-ping** option.

    > ./oradbver -c ~/tmp/db.json -ping
    
    Press Return to stop the pings...
    21:11:36  Inst: a01  Host: hasi  Service: SYS$USERS  Db.Name: a01  Db.Role: PRIMARY
    21:11:37  Inst: a01  Host: hasi  Service: SYS$USERS  Db.Name: a01  Db.Role: PRIMARY
    
### Running oradbver with **-ping-once** option.

    > ./oradbver -u andrej -p andrej -ping-once
    21:57:50  Inst: a01  Host: hasi  Service: SYS$USERS  Db.Name: a01  Db.Role: PRIMARY
    > echo $?
    0
    
    > ./oradbver -u andrej -p WrongPassword -ping-once
    21:58:19  database query error username="andrej" sid="" params={authMode:0 connectionClass:<nil> connectionClassLength:0 purity:0 newPassword:<nil> newPasswordLength:0 appContext:<nil> numAppContext:0 externalAuth:0 externalHandle:<nil> pool:<nil> tag:<nil> tagLength:0 matchAnyTag:0 outTag:<nil> outTagLength:0 outTagFound:0 shardingKeyColumns:<nil> numShardingKeyColumns:0 superShardingKeyColumns:<nil> numSuperShardingKeyColumns:0 outNewSession:0}: ORA-01017: invalid username/password; logon denied
    > echo $?
    1
