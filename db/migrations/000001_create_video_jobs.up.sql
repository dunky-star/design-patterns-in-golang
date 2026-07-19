CREATE TABLE `video_jobs` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `input_media_key` varchar(512) NOT NULL,
  `encoding_type` varchar(20) NOT NULL DEFAULT 'mp4',
  `status` varchar(20) NOT NULL DEFAULT 'pending',
  `output_reference` varchar(512) DEFAULT NULL,
  `error_message` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `video_jobs`
  (`input_media_key`, `encoding_type`, `status`)
VALUES
  ('puppy1.mp4', 'mp4', 'pending'),
  ('puppy2.mp4', 'mp4', 'pending');
