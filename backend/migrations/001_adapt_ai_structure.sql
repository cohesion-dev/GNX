-- Migration: Adapt database structure to AI interface (Page → Panel → Segment)
-- Date: 2025-10-25
-- Description: Restructure from flat storyboard to hierarchical page/panel/segment model

-- 1. Create comic_storyboard_page table
CREATE TABLE IF NOT EXISTS `comic_storyboard_page` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `section_id` INT UNSIGNED NOT NULL COMMENT '所属章节ID',
  `index` INT NOT NULL COMMENT '页码（从1开始）',
  
  -- AI interface aligned fields
  `image_prompt` TEXT NOT NULL COMMENT '整页图像提示词',
  `layout_hint` TEXT NOT NULL COMMENT '布局方式，如 2x2 grid',
  `page_summary` TEXT COMMENT '页面摘要',
  
  -- Business fields
  `status` VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending, completed, failed',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX `idx_section_page` (`section_id`, `index`),
  FOREIGN KEY (`section_id`) REFERENCES `comic_section`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='漫画章节分页表，对应AI StoryboardPage';

-- 2. Rename comic_storyboard to comic_storyboard_panel and restructure
-- First, create the new table
CREATE TABLE IF NOT EXISTS `comic_storyboard_panel` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `section_id` INT UNSIGNED NOT NULL COMMENT '所属章节ID',
  `page_id` INT UNSIGNED NOT NULL COMMENT '所属页面ID',
  `index` INT NOT NULL COMMENT '分格在页面中的索引（从1开始）',
  
  -- AI interface aligned fields
  `visual_prompt` TEXT NOT NULL COMMENT '分格视觉描述',
  `panel_summary` TEXT COMMENT '分格情节摘要',
  
  -- Business fields
  `image_url` VARCHAR(500) COMMENT '分格图片URL',
  `status` VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending, completed, failed',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX `idx_section_panel` (`section_id`, `index`),
  INDEX `idx_page_panel` (`page_id`, `index`),
  FOREIGN KEY (`section_id`) REFERENCES `comic_section`(`id`) ON DELETE CASCADE,
  FOREIGN KEY (`page_id`) REFERENCES `comic_storyboard_page`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='漫画分格表，对应AI StoryboardPanel';

-- 3. Rename comic_storyboard_detail to source_text_segment and restructure
CREATE TABLE IF NOT EXISTS `source_text_segment` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `panel_id` INT UNSIGNED NOT NULL COMMENT '所属分格ID',
  `index` INT NOT NULL COMMENT '在分格中的索引（从1开始）',
  
  -- AI interface aligned fields (complete field alignment)
  `text` TEXT NOT NULL COMMENT '语音文本片段',
  `character_refs` TEXT COMMENT '角色索引数组（JSON格式）',
  
  -- Business fields
  `tts_url` VARCHAR(500) COMMENT 'TTS音频URL',
  `role_id` INT UNSIGNED COMMENT '关联角色ID',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  INDEX `idx_panel_segment` (`panel_id`, `index`),
  INDEX `idx_role` (`role_id`),
  FOREIGN KEY (`panel_id`) REFERENCES `comic_storyboard_panel`(`id`) ON DELETE CASCADE,
  FOREIGN KEY (`role_id`) REFERENCES `comic_role`(`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='源文本片段表，对应AI SourceTextSegment';

-- 4. Drop old tables (after data migration if needed)
-- WARNING: This will delete all existing storyboard data!
-- Only execute after backing up data or confirming new structure is working
-- DROP TABLE IF EXISTS `comic_storyboard_detail`;
-- DROP TABLE IF EXISTS `comic_storyboard`;

-- 5. Data migration notes:
-- If you have existing data in comic_storyboard and comic_storyboard_detail,
-- you need to write custom migration logic to:
--   1. Create a page for each existing storyboard
--   2. Create a panel for each existing storyboard (1:1 mapping)
--   3. Migrate comic_storyboard_detail to source_text_segment
--
-- Example migration logic (commented out, customize as needed):
-- INSERT INTO comic_storyboard_page (section_id, `index`, image_prompt, layout_hint, image_url, status, created_at, updated_at)
-- SELECT section_id, `index`, image_prompt, 'single panel', image_url, status, created_at, updated_at
-- FROM comic_storyboard;
--
-- INSERT INTO comic_storyboard_panel (section_id, page_id, `index`, visual_prompt, status, created_at, updated_at)
-- SELECT s.section_id, p.id, 1, '', s.status, s.created_at, s.updated_at
-- FROM comic_storyboard s
-- JOIN comic_storyboard_page p ON s.section_id = p.section_id AND s.`index` = p.`index`;
--
-- INSERT INTO source_text_segment (panel_id, `index`, text, voice_name, voice_type, speed_ratio, is_narration, tts_url, created_at, updated_at)
-- SELECT panel.id, d.`index`, d.text, d.voice_name, d.voice_type, d.speed_ratio, d.is_narration, d.tts_url, d.created_at, d.updated_at
-- FROM comic_storyboard_detail d
-- JOIN comic_storyboard s ON d.storyboard_id = s.id
-- JOIN comic_storyboard_page p ON s.section_id = p.section_id AND s.`index` = p.`index`
-- JOIN comic_storyboard_panel panel ON panel.page_id = p.id;
