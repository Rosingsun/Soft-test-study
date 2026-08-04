
DROP TEMPORARY TABLE IF EXISTS q_rownum;
DROP TEMPORARY TABLE IF EXISTS c_rownum;
CREATE TEMPORARY TABLE q_rownum (id BIGINT PRIMARY KEY, rn INT NOT NULL, k VARCHAR(64));
SET @k='',@rn=0;
INSERT INTO q_rownum (id,rn,k) SELECT id, IF(@k=CONCAT(subject_id,':',sub_subject_id),@rn:=@rn+1,@rn:=1), @k:=CONCAT(subject_id,':',sub_subject_id) FROM questions WHERE chapter_id=0 AND status=1 ORDER BY subject_id, sub_subject_id, id;
CREATE TEMPORARY TABLE c_rownum (id BIGINT PRIMARY KEY, cidx INT NOT NULL, cnt INT NOT NULL, k VARCHAR(64));
SET @c='',@ci=0;
INSERT INTO c_rownum (id,cidx,cnt,k) SELECT id, IF(@c=CONCAT(subject_id,':',sub_subject_id),@ci:=@ci+1,@ci:=0), 0, @c:=CONCAT(subject_id,':',sub_subject_id) FROM chapters ORDER BY subject_id, sub_subject_id, id;
UPDATE c_rownum cr JOIN (SELECT c.id, COUNT(*) cnt FROM chapters c GROUP BY c.subject_id, c.sub_subject_id) s ON s.id=cr.id SET cr.cnt=s.cnt;
SELECT 'qrows', COUNT(*) FROM q_rownum;
SELECT 'crows', COUNT(*) FROM c_rownum;
SELECT 'badcnt', COUNT(*) FROM c_rownum WHERE cnt=0;
SELECT 'grp', k, cidx, cnt FROM c_rownum WHERE k IN ('6:23','8:1') ORDER BY k, cidx;
UPDATE questions q JOIN q_rownum qr ON qr.id=q.id JOIN chapters c ON c.subject_id=q.subject_id AND c.sub_subject_id=q.sub_subject_id JOIN c_rownum cr ON cr.id=c.id AND cr.cidx=(qr.rn-1)%cr.cnt SET q.chapter_id=c.id;
SELECT 'left', COUNT(*) FROM questions WHERE chapter_id=0;
