CREATE TABLE xdev_site."User"
(
  id SERIAL NOT NULL,
  username TEXT NOT NULL,
  fullname TEXT NOT NULL,
  passwordhash TEXT NOT NULL,
  email TEXT NOT NULL,
  isdisabled BOOLEAN NOT NULL
);
CREATE UNIQUE INDEX User_id_uindex ON xdev_site."User" (id);
CREATE UNIQUE INDEX User_username_uindex ON xdev_site."User" (username);
CREATE UNIQUE INDEX User_email_uindex ON xdev_site."User" (email);


CREATE TABLE xdev_site.calendartasks
(
  id SERIAL PRIMARY KEY NOT NULL,
  userid INT,
  taskid SERIAL NOT NULL,
  taskcaption TEXT,
  tto BIGINT,
  tfrom BIGINT,
  taskdescription TEXT,
  CONSTRAINT calendartasks_User_id_fk FOREIGN KEY (userid) REFERENCES "User" (id)
);
CREATE UNIQUE INDEX calendartasks_taskid_uindex ON xdev_site.calendartasks (taskid);
CREATE UNIQUE INDEX calendartasks_id_uindex ON xdev_site.calendartasks (id);
CREATE UNIQUE INDEX calendartasks_userid_uindex ON xdev_site.calendartasks (userid);