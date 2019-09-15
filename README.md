# oradbver

Use this Go program to get the Oracle Database version in JSON format. Additionally the 
parameters **-ping** or **-ping-once** ping the specified database every 1 second or just once.

 ## Working with oradbver
 
 ### Before starting oradver
 
 oradbver uses the Oracle OCI client library. That means, that either Oracle database software or 
 the client software must be installed. The environment parameters must be set as well. 
 As a test the SQL*Plus must run without any problems.
 
 This Go program is created and tested on Oracle Linux 7 64bit using Oracle Database Release 11.2. 
 This means, that the uploaded executable **oradbver** can be immediately downloaded and used for the 
 similar environments without any own build actions. 
 
 ### Parameters
 
 - **-u** database user. You can omit this parameter to use the "internal" connect.
 - **-p** database user password. The password will be asked, if you have specified the 
 database user.
 - **-d** database. This could be easy connect string or a TNS alias.
 - **-r** system privilege. At the moment only "sysdba" is supported.
 - **-c** JSON configuration file. You can specify all corresponding parameters in this file.
 - **-ping** Start pinging the database every second. You can stop this by pressing Return key.
 - **-ping-once** Ping the database once and exit with usual UNIX (0 or 1) return code.
 - **-short** Add more details to regular output (V$DATABASE, GV$INSTANCE).
 - **-full** Add more details to short output (DBA_REGISTRY, V$CONTAINERS).
  - **-pretty** Print pretty JSON output. 
    
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

### Running oradbver with **-short** and **-pretty** options.
If you want to get some columns from V$DATABASE and GV$INSTANCE views, you should use **-short** option.

    > ./oradbver -short
    {"Details":{"Release":"11.2.0.4","Version":11,"RAC":false,"CDB":false,"Role":"PRIMARY"},"Database":{"OpenMode":"READ WRITE","FlashbackOn":"RESTORE POINT ONLY","ForceLogging":"NO","ControlfileType":"CURRENT","ProtectionMode":"","ProtectionLevel":"MAXIMUM PERFORMANCE","SwitchoverStatus":"NOT ALLOWED","DataGuardBroker":"DISABLED"},"Instances":[{"InstanceNumber":1,"InstanceName":"a01","HostName":"hasi","Status":"OPEN","Parallel":"NO","ThreadNumber":1}]}

The output above is not very friendly. If you want to get this better formatted, you should specify
**-pretty** option.
    
    > ./oradbver -short -pretty
    {
        "Details": {
            "Release": "11.2.0.4",
            "Version": 11,
            "RAC": false,
            "CDB": false,
            "Role": "PRIMARY"
        },
        "Database": {
            "OpenMode": "READ WRITE",
            "FlashbackOn": "RESTORE POINT ONLY",
            "ForceLogging": "NO",
            "ControlfileType": "CURRENT",
            "ProtectionMode": "",
            "ProtectionLevel": "MAXIMUM PERFORMANCE",
            "SwitchoverStatus": "NOT ALLOWED",
            "DataGuardBroker": "DISABLED"
        },
        "Instances": [
            {
                "InstanceNumber": 1,
                "InstanceName": "a01",
                "HostName": "hasi",
                "Status": "OPEN",
                "Parallel": "NO",
                "ThreadNumber": 1
            }
        ]
    }
     
### Running oradbver with **-full** option.
    > ./oradbver -full        
    {"Details":{"Release":"12.2.0.1","Version":12,"RAC":false,"CDB":true,"Role":"PRIMARY"},"Database":{"OpenMode":"READ WRITE","FlashbackOn":"NO","ForceLogging":"NO","ControlfileType":"CURRENT","ProtectionMode":"","ProtectionLevel":"MAXIMUM PERFORMANCE","SwitchoverStatus":"NOT ALLOWED","DataGuardBroker":"DISABLED"},"Instances":[{"InstanceNumber":1,"InstanceName":"fcdb1","HostName":"hasi","Status":"OPEN","Parallel":"NO","ThreadNumber":1}],"Registry":[{"Name":"JServer JAVA Virtual Machine","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"OLAP Analytic Workspace","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle Database Catalog Views","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle Database Java Packages","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle Database Packages and Types","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle Database Vault","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle Label Security","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle Multimedia","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle OLAP API","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle Real Application Clusters","Version":"12.2.0.1.0","Status":"OPTION OFF"},{"Name":"Oracle Text","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle Workspace Manager","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle XDK","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Oracle XML Database","Version":"12.2.0.1.0","Status":"VALID"},{"Name":"Spatial","Version":"12.2.0.1.0","Status":"VALID"}],"Containers":[{"Name":"CDB$ROOT","ID":1,"OpenMode":"READ WRITE"},{"Name":"FCDB1_PDB1","ID":3,"OpenMode":"READ WRITE"},{"Name":"FCDB1_PDB2","ID":4,"OpenMode":"READ WRITE"},{"Name":"FCDB1_PDB3","ID":6,"OpenMode":"READ WRITE"},{"Name":"PDB$SEED","ID":2,"OpenMode":"READ ONLY"}]}

### Using Linux tool jq to parse JSON output from oradbver.
Usually you would use **-short** and **-full** options to get all details and pipe then to **jq**. This 
allows you to select specific database details without using SQL*Plus. These are some examples of
using **jq**:

Get database release only:

    > ./oradbver | jq '.Release'
    "12.2.0.1"
    
Get database open mode from the the short output:

    > ./oradbver -short | jq '.Database.OpenMode'
    "READ WRITE"

Get PDB names from the full output:
    
    > ./oradbver -full | jq '.Containers[].Name'
    "CDB$ROOT"
    "FCDB1_PDB1"
    "FCDB1_PDB2"
    "FCDB1_PDB3"
    "PDB$SEED"
