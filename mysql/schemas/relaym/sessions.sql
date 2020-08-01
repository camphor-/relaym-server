CREATE TABLE IF NOT EXISTS `sessions` (
  `id` VARCHAR(255) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL COMMENT 'セッションID（不変）',
  `name` VARCHAR(255) NOT NULL COMMENT 'Sessionの名前（不変）',
  `creator_id` VARCHAR(255) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL COMMENT 'sessionの作成者のユーザーID（不変）',
  `queue_head` INT NOT NULL COMMENT 'プレイヤーにセットされている曲のindex（0-indexed）（可変）',
  `state_type` ENUM('PLAY','PAUSE','STOP','ARCHIVED') NOT NULL,
  `device_id` varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '再生する端末のID(存在しない場合は空文字列)',
  `expired_at` datetime NOT NULL,
  `allow_to_control_by_others` TINYINT(1) NOT NULL DEFAULT '0',
  `progress_when_paused` INT NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  INDEX `sessions_user_id_fk_idx` (`creator_id` ASC) VISIBLE,
  CONSTRAINT `sessions_user_id_fk`
    FOREIGN KEY (`creator_id`)
    REFERENCES `users` (`id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE)
ENGINE = InnoDB;
