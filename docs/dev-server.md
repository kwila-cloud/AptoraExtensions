# Development Server

Information for setting up a development SQL server.

## Install Database Client

I used the community edition of [DbGate](https://www.dbgate.io/download-community/)

You will need this while setting up the SQL server, so go ahead and install the client first.

## Generate Seed Data

Do a back up on your Aptora SQL server - this should give you a `.bak` file.

## Set Up ProxMox Container for SQL Server

Run this [script](https://community-scripts.github.io/ProxmoxVE/scripts?id=sqlserver2022&category=Databases) in the shell on a ProxMox host.

### Configuration

- Send diagnostics: No
- Default settings
  - This will create a container with 10 GB disk and 2 GB RAM
- Automatically mount all available VAAPI devices?: Yes
- Do you want to run the SQL server setup now?: Yes
- Choose an edition: 2 (Developer)
- Enter administrator password - save in password manager

### Static IP

Set static IP address for LXC container in Network configuration.

## Restore Seed Data

### Transfer Backup File

You need to get the `.bak` file into the LXC container's file system somehow.

There are multiple approaches that could work for this. 

I used these steps:
- `scp` file from laptop to `/root` on ProxMox host
- From the ProxMox host:
  - `pct exec <CONTAINER-ID> mkdir /var/opt/mssql/backups`
  - `pct push <CONTAINER-ID> /root/<FILE-NAME>.bak /var/opt/mssql/backups/<FILE-NAME>.bak`

### Restore Backup File

Access your new dev server from your DB client.

Then run:
```
RESTORE FILELISTONLY
FROM DISK = '/var/opt/mssql/backups/<FILE-NAME>.bak';
```

Note the `LogicalName` values.

Then run the following:

(but replace `LogicalDataName`, `LogicalLogName`, and `MyDatabase` with correct values for your situation)

```
RESTORE DATABASE MyDatabase
FROM DISK = '/var/opt/mssql/backups/<FILE-NAME>.bak'
WITH 
    MOVE 'LogicalDataName' TO '/var/opt/mssql/data/MyDatabase.mdf',
    MOVE 'LogicalLogName' TO '/var/opt/mssql/data/MyDatabase_log.ldf',
    REPLACE;
```

Refresh the connection list in your DB client.

Verify you can access the restored DB.
