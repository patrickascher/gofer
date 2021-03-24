CREATE TABLE `users`
(
    `id`                int(11) unsigned       NOT NULL AUTO_INCREMENT,
    `login`             varchar(250)           NOT NULL DEFAULT '',
    `salutation`        enum ('MALE','FEMALE') NOT NULL,
    `name`              varchar(250)                    DEFAULT NULL,
    `surname`           varchar(250)                    DEFAULT NULL,
    `email`           varchar(250)              NOT NULL DEFAULT '',
    `state`             varchar(250)           NOT NULL DEFAULT '',
    `last_login`        datetime                        DEFAULT NULL,
    `failed_logins`     tinyint(11)                     DEFAULT NULL,
    `last_failed_login` datetime                        DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;


CREATE TABLE `roles`
(
    `id`          int(11) unsigned NOT NULL AUTO_INCREMENT,
    `name`        varchar(250)     NOT NULL DEFAULT '',
    `description` varchar(250)              DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE `role_roles`
(
    `role_id`  int(11) unsigned NOT NULL,
    `child_id` int(11) unsigned NOT NULL,
    KEY `role_id` (`role_id`),
    KEY `child_id` (`child_id`),
    CONSTRAINT `role_roles_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`),
    CONSTRAINT `role_roles_ibfk_2` FOREIGN KEY (`child_id`) REFERENCES `roles` (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;


CREATE TABLE `user_roles`
(
    `user_id` int(11) unsigned NOT NULL,
    `role_id` int(11) unsigned NOT NULL,
    KEY `user_id` (`user_id`),
    KEY `role_id` (`role_id`),
    CONSTRAINT `user_roles_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
    CONSTRAINT `user_roles_ibfk_2` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;


CREATE TABLE `user_protocols`
(
    `id`         int(11) unsigned NOT NULL AUTO_INCREMENT,
    `user_id`    int(10) unsigned NOT NULL,
    `key`        varchar(250)     NOT NULL DEFAULT '',
    `value`      varchar(250)              DEFAULT NULL,
    `created_at` datetime                  DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `user_options`
(
    `id`      int(11) unsigned NOT NULL AUTO_INCREMENT,
    `user_id` int(10) unsigned NOT NULL,
    `key`     varchar(250)     NOT NULL DEFAULT '',
    `value`   varchar(250)     NOT NULL DEFAULT '',
    `hide`    tinyint(1)       NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;


CREATE TABLE `routes`
(
    `id`         int(11) unsigned                                           NOT NULL AUTO_INCREMENT,
    `name`       varchar(250)                                               NOT NULL DEFAULT '',
    `pattern`    varchar(250)                                               NOT NULL DEFAULT '',
    `public`     tinyint(1)                                                 NOT NULL DEFAULT '0',
    `frontend`   tinyint(1)                                                 NOT NULL DEFAULT '0',
    `created_at` datetime                                                            DEFAULT NULL,
    `updated_at` datetime                                                            DEFAULT NULL,
    `deleted_at` datetime                                                            DEFAULT NULL,
    `method`     set ('DELETE','GET','HEAD','OPTIONS','PATCH','POST','PUT') NOT NULL DEFAULT '',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;


CREATE TABLE `refresh_tokens`
(
    `id`      int(10) unsigned NOT NULL AUTO_INCREMENT,
    `token`   varchar(250)     NOT NULL DEFAULT '',
    `user_id` int(11) unsigned NOT NULL,
    `expire`  datetime         NOT NULL,
    PRIMARY KEY (`id`),
    KEY `user_id` (`user_id`),
    CONSTRAINT `refresh_tokens_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;


CREATE TABLE `role_routes` (
                               `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
                               `role_id` int(11) unsigned NOT NULL,
                               `route_id` int(10) unsigned NOT NULL,
                               `route_type` varchar(100) NOT NULL DEFAULT '',
                               PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=39 DEFAULT CHARSET=utf8mb4;
