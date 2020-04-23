CREATE TABLE `spotify_auth` (
  `user_id` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT 'ユーザID',
  `access_token` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT 'Spotify OAuth2のアクセストークン',
  `refresh_token` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT 'Spotify OAuth2のリフレッシュトークン',
  `expiry` datetime NOT NULL COMMENT 'アクセストークンの有効期限',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
