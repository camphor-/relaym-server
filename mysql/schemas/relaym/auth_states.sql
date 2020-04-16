CREATE TABLE `auth_states` (
  `state` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT 'state',
  `redirect_url` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT 'OAuthが成功したときにリダイレクトするURL',
  UNIQUE KEY `state_state_uindex` (`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='SpotifyのOAuthに使う一時的なstate';
