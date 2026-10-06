// Package saleslead is an optional legacy downstream adapter retained for API
// compatibility. It is not the canonical responsibility of the houjin SDK; new
// sales eligibility, scoring and lead workflow belong to consumer applications.
//
// Package saleslead は houjin.CompanyProfile を 営業リスト DTO (SalesProspect) に
// 変換する純粋関数群を提供する。
//
// 案件 (Case) ベース対応:
//
//   - 1 つの 案件 (Case) には対象都道府県・市区町村・法人種別などの絞り込み条件が
//     ぶら下がる
//   - 案件に対して houjin.CompanyProfile の slice を流し込むと、案件条件で
//     フィルタしつつ SalesProspect の slice に変換する (BuildList)
//   - 各 SalesProspect は紐付く CaseID を保持するため、複数案件を 1 つのテーブルで
//     管理する DB 構造でもそのまま投入できる
//
// この package は DB 永続化や HTTP I/O を行わない。利用側がリポジトリ層を実装する。
package saleslead

import (
	"strings"

	houjin "github.com/sibukixxx/ja-company-base"
)

// ProspectStatus は営業リスト 1 件の進捗ステータス。
type ProspectStatus string

const (
	ProspectStatusNew        ProspectStatus = "new"         // 新規 (未接触)
	ProspectStatusContacted  ProspectStatus = "contacted"   // 接触済
	ProspectStatusInProgress ProspectStatus = "in_progress" // 商談中
	ProspectStatusWon        ProspectStatus = "won"         // 受注
	ProspectStatusLost       ProspectStatus = "lost"        // 失注
	ProspectStatusExcluded   ProspectStatus = "excluded"    // 対象外 (閉鎖法人など)
)

// Case は営業案件。1 つの案件は対象法人を絞り込む条件と、案件オーナーなどの
// メタ情報を持つ。
//
// PrefectureCodes / CityCodes / CompanyKinds はいずれもゼロ値の場合「絞り込み
// なし」として扱われ、すべての法人がマッチ対象となる。複数指定した場合は OR。
type Case struct {
	// ID は案件の一意識別子。利用側で UUID などを採番する。
	ID string

	// Name は案件名 (例: "2025Q2 関東 SaaS 新規開拓")。
	Name string

	// OwnerUserID は案件オーナーの社員 ID。任意。
	OwnerUserID string

	// PrefectureCodes は対象都道府県コード (総務省コード, 例: "13" = 東京都)。
	// 空 slice の場合は全都道府県が対象。
	PrefectureCodes []string

	// CityCodes は対象市区町村コード。空 slice の場合は全市区町村。
	// PrefectureCodes と組み合わせる場合、CityCodes は PrefectureCodes 配下の
	// ものだけを指定する想定 (利用側責務)。
	CityCodes []string

	// CompanyKinds は対象法人種別コード (例: "301" = 株式会社)。空 slice なら全種別。
	CompanyKinds []string

	// IncludeClosed が true の場合、閉鎖法人 (CloseDate が入っているもの) も
	// SalesProspect に含めるが、Status は ProspectStatusExcluded になる。
	// false の場合は BuildList の段階で除外される。
	IncludeClosed bool
}

// SalesProspect は営業リスト 1 件分の DTO。
//
// houjin.CompanyProfile の Source は「データのソース (国税庁 v4)」を保持する一方、
// SalesProspect の CaseID は「どの案件に紐付いた行か」を示す。
type SalesProspect struct {
	// CaseID は紐付く案件 ID (Case.ID をコピーしたもの)。
	CaseID string

	CorporateNumber string

	Name     string
	NameKana string

	Address houjin.Address

	Kind houjin.CompanyKind

	// Closed は元の CompanyProfile.CloseDate が入っていた場合 true。
	Closed bool

	// Status は新規生成時 ProspectStatusNew (Closed=true なら ProspectStatusExcluded)。
	// 営業活動の進行に応じて利用側が更新する。
	Status ProspectStatus

	Source houjin.SourceInfo
}

// Matches は profile が case_ の絞り込み条件にマッチするかを判定する。
//
//   - PrefectureCodes が空でなければ、都道府県名から算出された PrefectureCode が一致する必要がある
//     ただし本 SDK の Address には PrefectureCode が含まれないため、判定は呼び出し側が
//     事前にコード変換してから渡すか、CompanyProfile が houjin の検索段階で
//     PrefectureCode フィルタ済みであることを期待する。
//     この関数では PrefectureName が空でないことだけを最低限チェックする。
//   - CompanyKinds が空でなければ、profile.Kind と前方一致または完全一致で比較する。
//   - IncludeClosed が false かつ profile が閉鎖されていれば false を返す。
func (c Case) Matches(profile houjin.CompanyProfile) bool {
	if !c.IncludeClosed && isClosed(profile) {
		return false
	}
	if len(c.CompanyKinds) > 0 && !containsString(c.CompanyKinds, profile.Kind.String()) {
		return false
	}
	return true
}

// FromCompanyProfile は houjin.CompanyProfile を SalesProspect に変換する。
// case_ の ID と IncludeClosed の扱いを反映するが、case_ の絞り込み条件
// (PrefectureCodes など) のチェックは行わない。BuildList を使うとフィルタも
// 同時に行われる。
func FromCompanyProfile(case_ Case, profile houjin.CompanyProfile) SalesProspect {
	closed := isClosed(profile)
	status := ProspectStatusNew
	if closed {
		status = ProspectStatusExcluded
	}
	return SalesProspect{
		CaseID:          case_.ID,
		CorporateNumber: profile.CorporateNumber,
		Name:            profile.Name,
		NameKana:        profile.NameKana,
		Address:         profile.Address,
		Kind:            profile.Kind,
		Closed:          closed,
		Status:          status,
		Source:          profile.Source,
	}
}

// BuildList は profiles の中から case_ にマッチするものだけを SalesProspect に
// 変換して返す。順序は profiles の順序を保つ。
// 同一法人番号の重複は最初に現れたものを採用する (de-dup)。
// IncludeClosed=false の場合、閉鎖法人は除外される。
func BuildList(case_ Case, profiles []houjin.CompanyProfile) []SalesProspect {
	out := make([]SalesProspect, 0, len(profiles))
	seen := make(map[string]struct{}, len(profiles))
	for _, p := range profiles {
		if !case_.Matches(p) {
			continue
		}
		if p.CorporateNumber != "" {
			if _, ok := seen[p.CorporateNumber]; ok {
				continue
			}
			seen[p.CorporateNumber] = struct{}{}
		}
		out = append(out, FromCompanyProfile(case_, p))
	}
	return out
}

func isClosed(p houjin.CompanyProfile) bool {
	return strings.TrimSpace(p.CloseDate) != ""
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
