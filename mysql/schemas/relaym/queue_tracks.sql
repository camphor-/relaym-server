CREATE TABLE IF NOT EXISTS `queue_tracks` (
  `id` VARCHAR(255) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL COMMENT 'ID（不変）',
  `index` INT NOT NULL COMMENT 'session内でのindex（0-indexed）（不変）',
  `uri` VARCHAR(255) NOT NULL COMMENT 'Spotify APIから返ってくるuri（不変）',
  `session_id` VARCHAR(255) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `queue_tracks_index_session_id_uindex` (`session_id` ASC, `index` ASC) VISIBLE,
  CONSTRAINT `tracks_session_id_fk`
    FOREIGN KEY (`session_id`)
    REFERENCES `sessions` (`id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE)
ENGINE = InnoDB;
