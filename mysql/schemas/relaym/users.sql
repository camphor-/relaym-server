CREATE TABLE `users` (
  `id` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT 'ユーザID (不変)',
  `spotify_user_id` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT 'SpotifyのユーザID (不変)',
  `display_name` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '表示名 (変更可能)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `users_spotify_user_id_uindex` (`spotify_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
