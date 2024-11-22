CREATE TABLE `channels` (
  `id` integer PRIMARY KEY,
  `channel_id` varchar(255),
  `channel_name` varchar(255),
  `created_at` timestamp
);

CREATE TABLE `videos` (
  `id` integer PRIMARY KEY, 
  `title` varchar(255),
  `upload_date` date,
  `url` varchar(255),
  `thumbnail` varchar(255),
  `transcript` longtext,
  `description` text,
  `video_id` varchar(255),
  `channel_id` integer,
  `channel_name` varchar(255),
  `video_text_data` longtext,
  `created_at` timestamp default current_timestamp,
  FOREIGN KEY (`channel_id`) REFERENCES `channels` (`channel_id`)
);
