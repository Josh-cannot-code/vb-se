CREATE TABLE `channels` (
  `id` integer PRIMARY KEY,
  `channel_id` varchar(255),
  `channel_name` varchar(255),
  `created_at` timestamp default current_timestamp
);

CREATE TABLE `videos` (
  `id` integer PRIMARY KEY, 
  `title` varchar(255),
  `upload_date` date,
  `url` varchar(255),
  `thumbnail` varchar(255),
  `transcript` text,
  `description` text,
  `video_id` varchar(255),
  `channel_id` integer,
  `channel_name` varchar(255),
  `created_at` timestamp default current_timestamp,
  FOREIGN KEY (`channel_id`) REFERENCES `channels` (`channel_id`)
);

CREATE VIRTUAL TABLE `video_text_data`
USING FTS5(video_id,title,description,transcript);
