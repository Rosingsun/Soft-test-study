-- =============================================================
-- 000005_checkin_rank_backfill.up.sql
-- 回填 check_ins 历史数据的「首次打卡」快照（rank_*）：
-- 历史行 rank_* 为 0 时，将其回填为当日 correct_count/accuracy/duration，
-- 避免排名与个人平均正确率被清零。
-- =============================================================

UPDATE `check_ins`
   SET `rank_correct_count` = `correct_count`,
       `rank_accuracy`      = `accuracy`,
       `rank_duration`      = `duration`
 WHERE `rank_correct_count` = 0
   AND `rank_accuracy` = 0;
