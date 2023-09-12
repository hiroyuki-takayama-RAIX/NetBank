# NetBank
「golang 最初の手順」のオンライン銀行プロジェクトのプログラムを、学習のために改良しています。

### 履歴
[2023/09/08] 

・docker-composeを使用してpostgreSQLを起動

・container上のDBにmain.goからアクセス

・coreモジュールにビジネスロジックのコードを記載

・coreモジュールのテストコードを記載、すべてのテストに成功

・apiモジュールにAPIを実装

・apiモジュールのテストコードを記載、withdraw関数が成功した場合のテストに失敗

[2023/09/21までの課題]

・パスパラメーター、クエリパラメーター,ボディパラメーターを使い分ける

・　GET、POST、PUT、PATCH、DELETEのリクエストを使い分ける

・GETを使用して全体検索のAPIを作成する

・withdrawのテストが成功するように修正する

・postmanを使用して結合テストを行う

・main.goもコンテナ上で起動するようにする
