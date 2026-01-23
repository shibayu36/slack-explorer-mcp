package main

import (
	"fmt"
	"testing"
)

func TestConvertHTMLToMarkdown_RealCanvas(t *testing.T) {
	htmlContent := `<h1 id='temp:C:BVF3da2ce54278c4415a93435dc1'>新メンバーのオンボーディング(shibayu36テスト)</h1>

<p id='BVF9CAEgcSN' class='line'>[name]さん、ようこそ！Slack 仲間が増えてうれしいです！</p>

<h1 id='BVF9CA0e4Ou'>✅ 第 1 週のタスク</h1>

<p id='BVF9CAomgLn' class='line'>今週中に、以下のタスクを完了しましょう。</p>

<div data-section-style='7' class="" style=""><ul id='BVF9CAMrAKK'><li id='BVF9CASJ1YB' class='' style='' value='1'><span id='BVF9CASJ1YB'>経時的に進捗を記録できるようチェックリスト形式を使用します。</span>

<br/></li></ul></div><h1 id='BVF9CAt8XIh'>📅 参加するミーティング</h1>

<p id='BVF9CA3wBTm' class='line'>最初の数週間に参加するセッションとイベントの概要です。</p>

<table><tr><td><p id='BVF9CA2TXdW' class='line'>日付</p>

</td><td><p id='BVF9CAHxft3' class='line'>時間</p>

</td><td><p id='BVF9CAzRdYt' class='line'>イベント</p>

</td></tr><tr><td><p id='BVF9CA53Fm4' class='line'>日付を追加する</p>

</td><td><p id='BVF9CArlm1k' class='line'>時間を追加する</p>

</td><td><p id='BVF9CAWb7wV' class='line'>研修セッション、チームの定例会議、1 対 1 の面談などのイベントを追加する</p>

</td></tr><tr><td><p id='temp:C:BVF9e25b92be92649e59cde32c3e' class='line'>​</p>

</td><td><p id='temp:C:BVF5674017dbbf843db828026edc' class='line'>​</p>

</td><td><p id='temp:C:BVF8cf7f56f2ebd4829a950c7bee' class='line'>​</p>

</td></tr></table>
<h1 id='BVF9CArOLlP'>👥 サポートしてくれる先輩たち</h1>

<p id='BVF9CAnTSU4' class='line'>研修をサポートする先輩が、プロセスや定期的なミーティングについての質問に答えてくれます。</p>

<p id='BVF9CAj5cvY' class='line'>@ を使用してチームのメンバーをメンションします。それからメンバーの表示名を右クリックしてカード形式に変更します。</p>

<h1 id='BVF9CArfcEN'>📺 チャンネルへの参加</h1>

<p id='BVF9CA5uKaI' class='line'>チーム用、プロジェクト用、交流用など複数のチャンネルが用意されています。ワークフローをクリックすると、そのすべてに参加できます。</p>


<p class='embedded-link'>Link URL: https://slack.com/shortcuts/Ft1234567890/abcdef1234567890</p>
<h1 id='BVF9CAXyPuA'>📚 企業文化についての資料を読む</h1>

<h1 id='BVF9CAhCJZI'>💡 このような canvas を作るためのヒント</h1>

<p id='BVF9CA3eFCj' class='line'>1️⃣ @ を使ってチームメンバーをタグづけできます</p>

<p id='BVF9CA02ows' class='line'>2️⃣ ツールバーのチェックリストアイコン ✅ を使って、実施項目のリストを作成できます</p>

<p id='BVF9CAkJ6G3' class='line'>3️⃣ Slack や YouTube からリンク 🔗 をコピーして canvas にペーストすると、特別なカードに変わります</p>

<p id='temp:C:BVFe2fc99b752894b1995741b93c' class='line'>​</p>

<h1 id='temp:C:BVFfb55076f644f434a8b4883dde'>見出し1</h1>

<div data-section-style='5' class="" style=""><ul id='temp:C:BVF371a5399bd774ce79dc0e37ff'><li id='temp:C:BVFfdab15daa4fc4249a28d81ef6' class='parent' style='' value='1'><span id='temp:C:BVFfdab15daa4fc4249a28d81ef6'>リスト1</span>

<br/></li><ul><li id='temp:C:BVFdd9388ae22214eef89931abd2' class='parent' style=''><span id='temp:C:BVFdd9388ae22214eef89931abd2'>リスト1-1</span>

<br/></li><ul><li id='temp:C:BVF39b5c62f56f140f288111a3a9' class='' style=''><span id='temp:C:BVF39b5c62f56f140f288111a3a9'>リスト1-1-1</span>

<br/></li></ul><li id='temp:C:BVFa6064dd0740d415bbabebc7a4' class='' style=''><span id='temp:C:BVFa6064dd0740d415bbabebc7a4'>リスト1-2</span>

<br/></li></ul><li id='temp:C:BVFd068c336666d4d2a8cf5137c3' class='parent' style=''><span id='temp:C:BVFd068c336666d4d2a8cf5137c3'>リスト2</span>

<br/></li><ul><li id='temp:C:BVF03bfdb0ef71d432fad50767c2' class='' style=''><span id='temp:C:BVF03bfdb0ef71d432fad50767c2'>リスト2-1</span>

<br/></li></ul></ul></div><p id='temp:C:BVF1ba4256e6db34e8b96b0ac433' class='line'>​</p>

<h2 id='temp:C:BVF2105a1e3e073451badacb1e1a'>見出し2</h2>

<div data-section-style='6' class="" style=""><ul id='temp:C:BVF04287fddd09e4405ae30acb38'><li id='temp:C:BVF31a7e52d471f46e88e8d7b962' class='parent' style='' value='1'><span id='temp:C:BVF31a7e52d471f46e88e8d7b962'>ordered list1</span>

<br/></li><ul><li id='temp:C:BVF98e60e5c76a549868714d05e5' class='parent' style=''><span id='temp:C:BVF98e60e5c76a549868714d05e5'>orderd list1-1</span>

<br/></li><ul><li id='temp:C:BVF9947746e5ddd4e53899816c5e' class='' style=''><span id='temp:C:BVF9947746e5ddd4e53899816c5e'>ordered list1-1-1</span>

<br/></li></ul><li id='temp:C:BVF8e5cec1426d742a78a68ee27d' class='' style=''><span id='temp:C:BVF8e5cec1426d742a78a68ee27d'>ordered list1-2</span>

<br/></li></ul><li id='temp:C:BVF4871458942a34c43ae8cb3c42' class='parent' style=''><span id='temp:C:BVF4871458942a34c43ae8cb3c42'>ordered list2</span>

<br/></li><ul><li id='temp:C:BVFb50e62a6a262438582a33b8be' class='' style=''><span id='temp:C:BVFb50e62a6a262438582a33b8be'>ordered list2-1</span>

<br/></li></ul></ul></div><p id='temp:C:BVF7f0b46add0134dd0a25e97812' class='line'>​</p>

<h3 id='temp:C:BVF433bfa7f54394e0a8fba2219c'>見出し3</h3>

<div data-section-style='7' class="" style=""><ul id='temp:C:BVFe7650c0693094f33ac4f0576f'><li id='temp:C:BVFbbae60fed69f4e8283ecdd65c' class='parent' style='' value='1'><span id='temp:C:BVFbbae60fed69f4e8283ecdd65c'>check1</span>

<br/></li><ul><li id='temp:C:BVF779fd34d78bd4f89b759f93cc' class='parent' style=''><span id='temp:C:BVF779fd34d78bd4f89b759f93cc'>check1-1</span>

<br/></li><ul><li id='temp:C:BVFc1287cda77524bb7acd535922' class='' style=''><span id='temp:C:BVFc1287cda77524bb7acd535922'>check1-1-1</span>

<br/></li><li id='temp:C:BVF023810cd2613416c80534ed16' class='checked' style=''><span id='temp:C:BVF023810cd2613416c80534ed16'>check1-1-2</span>

<br/></li></ul><li id='temp:C:BVFf13fc27ad0814a4dabf27badf' class='' style=''><span id='temp:C:BVFf13fc27ad0814a4dabf27badf'>check1-2</span>

<br/></li></ul><li id='temp:C:BVFec7cbedc1b5a43e88474cc2ea' class='parent' style=''><span id='temp:C:BVFec7cbedc1b5a43e88474cc2ea'>check2</span>

<br/></li><ul><li id='temp:C:BVFa2a35bfa35af4f68afe583c7e' class='' style=''><span id='temp:C:BVFa2a35bfa35af4f68afe583c7e'>check-2-1</span>

<br/></li></ul></ul></div><p id='temp:C:BVF98357ece631e41708a7119239' class='line'>​</p>

<p id='temp:C:BVF83dbd0deb3374f3380876adf9' class='line'>column1</p>

<p id='temp:C:BVFed73d11cb2e34717bcaac117f' class='line'>カラムを二つの左側</p>

<p id='temp:C:BVFba58f9815754475e9ed32b119' class='line'>column2</p>

<p id='temp:C:BVF23a10d50111f4e05b20e9a345' class='line'>カラム2つの右側</p>

<p id='temp:C:BVF62549dd4d0694bdea492325fe' class='line'>カラム1</p>

<p id='temp:C:BVFddde9097375046a9a7cf23bf6' class='line'>カラム2</p>

<p id='temp:C:BVF06f6a5b0fb9d486f836832af2' class='line'>カラム3</p>

<p id='temp:C:BVFaa7e7ec24beb40e9a915e5704' class='line'>​</p>

<p id='temp:C:BVFaea7464864c24e4089828f420' class='line'>本文</p>

<p id='temp:C:BVFdbb371a465f545d2901c3c4e8' class='line'>コメントをつけたい</p>

<p id='temp:C:BVFec55a281c7a54cb0b2149305b' class='line'>リアクションをつけたい</p>

<p id='temp:C:BVF47e81562ae764960851231bde' class='line'>​</p>

<p id='temp:C:BVF36ab1aa9e7a44796aee7999aa' class='line'>ファイル添付</p>


<p class='embedded-file'>File ID: F1234567890 File URL: https://example.slack.com/files/U1234567890/F1234567890/sample_image.jpg</p>
<p id='temp:C:BVFdab52f3890484e1fbca28989e' class='line'>​</p>

<blockquote id='temp:C:BVF77b3e7775c184d3eb8b67d6e2'>ここは引用です<br>ここも引用です</blockquote>

<pre id='temp:C:BVF05fe6816ff2f4d458011e4167' class='prettyprint'>ここはコードブロックです<br>ここもコードブロックです</pre>

<p id='temp:C:BVF6cd72a6645a4475fb4945aec6' class='line'><b>太字</b></p>

<p id='temp:C:BVF21e2581f04714160954a56629' class='line'><i>斜体</i></p>

<p id='temp:C:BVF7a7a04661b4f458c84bc525a5' class='line'><u>下線</u></p>

<p id='temp:C:BVF65cc02a7d4cc46debd8a73bb1' class='line'><del>取り消し線</del></p>

<p id='temp:C:BVF48b7d9a3c6094e4389f1da179' class='line'>​</p>

<p id='temp:C:BVF4556a961fefb49bcb241dd4ef' class='line'><a href="https://example.com/">リンク</a></p>

<p id='temp:C:BVF96ba32f2599040f896baf62a4' class='line'>​</p>

<p id='temp:C:BVF70b66a1c70df4161921ed281b' class='line'>​</p>

<p id='temp:C:BVF48950c8f8aaf406c81449feac' class='line'>​</p>

<p id='temp:C:BVF51d328571ffe444785a70686d' class='line'>​</p>

`

	result := ConvertHTMLToMarkdown(htmlContent)
	fmt.Println("=== Converted Markdown ===")
	fmt.Println(result)
	fmt.Println("=== End ===")

	// Just run to see the output
	t.Log("Conversion completed - check output above")
}
