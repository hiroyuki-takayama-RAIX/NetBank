-- テーブルの作成
CREATE TABLE duplicate_table (
	id INT,
	name varchar(80)
);

-- PRIMARY KEY に設定することで、NULLではなく、重複する値が新しく挿入されることを許さないようにする。
CREATE TABLE customer (
    id INT PRIMARY KEY,
    username VARCHAR(255),
    addr VARCHAR(255),
    phone VARCHAR(53)
);


-- REFERENCE句によって他のテーブルを参照することによって異なるテーブルを参照できるようになる。
-- この場合、customer.id に存在しない値をaccount.idに挿入することはできない。
-- また、account.idに存在する値と同じcustomer.idを持つcustomerのレコードを削除できないようになる。
CREATE TABLE account (
  id INT PRIMARY KEY,
  balance FLOAT,
  FOREIGN KEY (id) REFERENCES customer(id)
);


-- テーブルにレコードを挿入する。
INSERT INTO customer (id, username, addr, phone) 
VALUES (1001, 'John', 'Los Angeles, California', '(213) 444 0147');

INSERT INTO account (id, balance) 
VALUES (1001, 100);

INSERT INTO customer (id, username, addr, phone) 
VALUES (3003, 'Ide Non No', 'Ta No Tsu', '(0120) 117 117');

INSERT INTO account (id, balance) 
VALUES (3003, 100);

-- カラム名を明示せずに挿入することもできる
INSERT INTO duplicate_table 
VALUES (1, 'a'), (2, 'b'), (3, 'c'), (3, 'd'), (2, 'c'), (4, 'a'), (1, 'a');


-- テスト用のデータを作成するときは以下のようにクエリを実行する。
-- generate_series(1, 1000)で、1から1000までの連続した値をテーブルに挿入している。
-- random()、を使用して、ランダムな数値を出力している。
-- ::textによって、random()の返値を文字列に変換している。
-- md5()によってrandom()::textの結果をハッシュ化して、ユーザー名としてテーブルに挿入している。
-- md5(random()::text)によって出力された最初から15番目までの文字を抽出し、addr列に格納している。
-- random()関数を使用して0から1未満のランダムな数値を生成し、それを1億未満の整数に変換している。
-- '+1-'という文字列と結合して、phone列に挿入する。
INSERT INTO customer (id, username, addr, phone)
SELECT
	generate_series(1, 1000) AS id,
	md5(random()::text) AS username,
	substr(md5(random()::text), 1, 15) AS addr,
	'+1-' || floor(random() * 1000000000)::bigint AS phone;


-- テーブルからレコードを取得する。
-- INNER JOIN (JOINだけでも同じ) によって二つのテーブルを結合させている。
-- テーブルの結合の際に、二つのテーブルに同じ列名が存在する場合、「テーブル名.カラム名」というようにどのテーブルのカラムなのか明確にする必要がある。
-- WHERE句に論理演算式を記入して、条件に該当するレコードのみを取得することができる。
SELECT account.id, balance, username, addr, phone
FROM account
INNER JOIN customer
ON account.id=customer.id
WHERE account.id=$1


-- レコードの更新
-- 必ずWHEREで一意の識別しで更新を行うレコードを指定すること。
UPDATE customer
SET username='unknown', addr="unknown", phone='unknown'
WHERE id=3003;

-- レコードの削除
-- 必ずWHEREで一意の識別しで更新を行うレコードを指定すること。
DELETE FROM account WHERE id=1001;

-- 以下のクエリはaccountテーブルのすべてのレコードを削除するので、絶対にやらないこと。
DELETE FROM account;

-- テーブルの削除
DROP TABLE account, customer;

-- SELECTにDISTINCTを加えることで、重複を削除してレコードを取得できる。
-- この場合だと２つ目の (1, 'a') のレコードを除くすべてのレコードが取得できる。
SELECT DISTINCT id, name
FROM duplicate_table
ORDER BY id, name;

-- DISTINCT ON の場合は重複を無くしたいカラムを指定しつつ、他のカラムも取得できるようにする。
-- この場合だと、(3, 'd'), (2, 'c'), (1, 'a') を除くレコードが取得できる。
SELECT DISTINCT ON (id) id, name
FROM duplicate_table
ORDER BY id, name;

-- JOIN系のn操作nためにテーブルを作成
CREATE TABLE right_table (
	id INT PRIMARY KEY,
	name varchar(80)
);

INSERT INTO right_table 
VALUES (1, 'a'), (2, 'b'), (3, 'c');
	
CREATE TABLE left_table (
	id INT PRIMARY KEY,
	name varchar(80)
);

INSERT INTO left_table 
VALUES (1, 'd'), (2, 'e'), (6, 'f');

-- LEFT OUTER JOIN
-- left_table のレコードをすべて出力し、right_table側にはnullが含まれる。
SELECT * 
FROM left_table
LEFT OUTER JOIN right_table
ON left_table.id=right_table.id
ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;


-- RIGHT OUTER JOIN
-- left_table のレコードをすべて出力し、left_table側にはnullが含まれる。
SELECT * 
FROM left_table
RIGHT OUTER JOIN right_table
ON left_table.id=right_table.id
ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;


-- FULL OUTER JOIN
-- LEFT JOIN を先に行って、その後でRIGHT JOINで発生するNULLを含むレコードを加える。
SELECT * 
FROM left_table
FULL OUTER JOIN right_table
ON left_table.id=right_table.id
ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;


-- USING
-- 二つテーブルに同じカラム名がある場合、USING()で同じ名前のカラムを使用してテーブルを結合する。
-- ON ~ の場合だと SELECT で取得するカラムを＊にするとidがカラムに二つ存在するようになる。USING()の場合は重複する名前のカラムは自動的に消去される。
-- 以下のクエリは`SELECT left_table.id, left_table.name, left_table.name FROM left_table LEFT OUTER JOIN right_table ON left_table.id=right_table.id ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;`と同じである。 
SELECT * 
FROM left_table
LEFT OUTER JOIN right_table
USING (id)
ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;


-- USING ALL COLUMN
-- USING() に複数のカラムを指定した場合、カラムのペアと一致するレコードによってテーブルを結合する。
-- 以下のクエリは`SELECT left_table.id, left_table.name FROM left_table LEFT OUTER JOIN right_table ON left_table.id=right_table.id AND left_table.name=right_table.name ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;`と同じである。  
SELECT * 
FROM left_table
LEFT OUTER JOIN right_table
USING (id, name)
ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;


-- NATURAL JOIN
-- 二つのテーブルから同じ名前のカラムを自動で検知して、テーブルの結合を行う。
SELECT * 
FROM left_table
NATURAL LEFT OUTER JOIN right_table
ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;

-- LIMIT
-- LIMIT句を使用することで、取得したレコードの先頭~行だけを抽出できる
SELECT id, name
FROM duplicate_table
ORDER BY id, name
LIMIT 3;


-- OFFSET
-- OFFSET句を使用することで、LIMITで抽出を始めるレコードの先頭行数を指定する。
SELECT id, name
FROM duplicate_table
ORDER BY id, name
LIMIT 3 OFFSET 2;


-- UNION
-- 1番目のクエリの2番目のクエリで取得できるレコードを足し合わせて、、重複部分を除いたレコードを取得する。
-- この場合だと (1, 'a'), (2, 'b'), (3, 'c') (3, 'd'), (2, 'c'), (4, 'a') を取得する。
-- UNION を UNION ALL に変更した場合は (1, 'a'), (1, 'a'), (1, 'a'), (2, 'b'), (2, 'b'), (3, 'c'), (3, 'c') (3, 'd'), (2, 'c'), (4, 'a') を取得する。
SELECT *
FROM duplicate_table
ORDER BY id, name
UNION
SELECT *
FROM right_table
ORDER BY id, name;


-- INTERSECT
-- 1番目のクエリと2番目のクエリで取得できるレコードで、共通するレコードのみを重複を除いて抽出する。
-- この場合だと(1, 'a'), (2, 'b'), (3, 'c')を取得する。
-- INTERSECT を INTERSECT ALL に変更した場合は (1, 'a'), (1, 'a'), (1, 'a'), (2, 'b'), (2, 'b'), (3, 'c'), (3, 'c')　を取得する。
SELECT *
FROM duplicate_table
ORDER BY id, name
INTERSECT
SELECT *
FROM right_table
ORDER BY id, name;


-- EXCEPT
-- 1番目のクエリに含まれていて、2番目のクエリに含まれていないすべてのレコードを重複なして取得する。
-- この場合だと(3, 'd'), (2, 'c'), (4, 'a') を。取得する。
-- EXCEPT を EXCEPT ALL に変更した場合は (3, 'd'), (2, 'c'), (4, 'a')　を取得する。
SELECT *
FROM duplicate_table
ORDER BY id, name
EXCEPT	
SELECT *
FROM right_table
ORDER BY id, name;


-- GROUP BY
-- 指定したカラムに含まれる項目に集約する。
-- GROUP BY に含まれていないカラムに関しては集約式を用いて参照を行う。
-- 例えば集約式のsum()を使用して、nameの各項目の合計値を算出できる。
-- この場合は ('a', 6), ('b', 2), ('c', 5), ('d', 3) が取得できる。
SELECT name, sum(id)
FROM duplicate_table
ORDER BY name, id
GROUP BY name;

-- HAVING
-- GROUP BY で集約した項目に、より詳細な条件をつけてレコードを取得する。
-- この場合だと('a', 6), ('c', 5),が取得できる。
SELECT name, sum(id)
FROM duplicate_table
ORDER BY name, id
GROUP BY name
HAVING name > 4;


-- GROUPING SET
-- GROUP BY に比べて、複数の項目の集約を一度で行うことができる。
-- この場合はproduct_idごとのsalesの合計値、department_idごとのsalesの合計値、product_idとdepartment_idを掛け合わせた項目ごとのsalesの３種類の項目でsalesの集約を行うことができる。
SELECT product_id, department_id, SUM(sales) AS total_sales
FROM sales
GROUP BY GROUPING SETS ((product_id), (department_id), (product_id, department_id));


-- CUBE
-- CUBE句内で指定されたカラムのすべての組み合わせを作成し、それぞれの項目ごとに集約を。行う。
-- この場合だと、yearとmonthとproduct_idの項目から、考えられるすべての組み合わせを算出し、それぞれの項目ごとにsalesを集約する。
SELECT
    EXTRACT(YEAR FROM sale_date) AS year,
    EXTRACT(MONTH FROM sale_date) AS month,
    product_id,
    SUM(sales) AS total_sales
FROM sales
GROUP BY CUBE (year, month, product_id);


-- ROLLUP
-- ROLLUP句内で指定されたカラムの中で、左側に設定したカラムを中心に項目を作成して、それぞれで集計を行う。
-- この場合だと、(year, month, product_id), (year, month), (year) ごとにsalesの集計を行う。
