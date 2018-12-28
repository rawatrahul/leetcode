package leetcode

import (
	"bytes"
	"fmt"
	"path"
	"strconv"
	"strings"
)

func GetTags() (tags []tagType) {
	data := fileGetContents("tag/tags.json")
	jsonDecode(data, &tags)
	return
}

func SaveTags(tags []tagType) {
	ts := GetTags()
	var flag = make(map[string]bool)
	for _, tag := range ts {
		flag[tag.Slug] = true
	}
	for _, tag := range tags {
		if !flag[tag.Slug] {
			ts = append(ts, tag)
		}
	}
	filePutContents("tag/tags.json", jsonEncode(ts))
}

func GetTopicTag(slug string) (tt topicTagType) {
	jsonStr := `{
		"operationName": "getTopicTag",
		"variables": {
		"slug": "` + slug + `"
		},
		"query": "query getTopicTag($slug: String!) {\n  topicTag(slug: $slug) {\n    name\n    translatedName\n    questions {\n      status\n      questionId\n      questionFrontendId\n      title\n      titleSlug\n      translatedTitle\n      stats\n      difficulty\n      isPaidOnly\n      topicTags {\n        name\n        translatedName\n        slug\n        __typename\n      }\n      companyTags {\n        name\n        translatedName\n        slug\n        __typename\n      }\n      __typename\n    }\n    frequencies\n    __typename\n  }\n  favoritesLists {\n    publicFavorites {\n      ...favoriteFields\n      __typename\n    }\n    privateFavorites {\n      ...favoriteFields\n      __typename\n    }\n    __typename\n  }\n}\n\nfragment favoriteFields on FavoriteNode {\n  idHash\n  id\n  name\n  isPublicFavorite\n  viewCount\n  creator\n  isWatched\n  questions {\n    questionId\n    title\n    titleSlug\n    __typename\n  }\n  __typename\n}\n"
	}`
	filename := "topic_tag_" + strings.Replace(slug, "-", "_", -1) + ".json"
	graphQLRequest(filename, jsonStr, &tt)
	return
}

type tagType struct {
	Name           string
	Slug           string
	TranslatedName string
}

type topicTagType struct {
	Errors []errorType `json:"errors"`
	Data   ttDataType  `json:"data"`
}

type ttDataType struct {
	TopicTag ttType `json:"topicTag"`
}

type ttType struct {
	Name           string           `json:"name"`
	TranslatedName string           `json:"translatedName"`
	Questions      []ttQuestionType `json:"questions"`
}

type ttQuestionType struct {
	QuestionId         string    `json:"questionId"`
	QuestionFrontendId string    `json:"questionFrontendId"`
	Title              string    `json:"title"`
	TitleSlug          string    `json:"titleSlug"`
	TranslatedTitle    string    `json:"translatedTitle"`
	TranslatedContent  string    `json:"translatedContent"`
	IsPaidOnly         bool      `json:"isPaidOnly"`
	Difficulty         string    `json:"difficulty"`
	TopicTags          []tagType `json:"topicTags"`
}

func (question ttQuestionType) TagsStr() string {
	var buf bytes.Buffer
	format := "[[%s](https://github.com/openset/leetcode/tree/master/tag/%s/README.md)] "
	for _, tag := range question.TopicTags {
		buf.WriteString(fmt.Sprintf(format, tag.ShowName(), tag.Slug))
	}
	SaveTags(question.TopicTags)
	return string(buf.Bytes())
}

func (tag tagType) SaveContents() {
	questions := GetTopicTag(tag.Slug).Data.TopicTag.Questions
	var buf bytes.Buffer
	buf.WriteString("<!--|This file generated by command(leetcode tag); DO NOT EDIT.            |-->")
	buf.WriteString(authInfo)
	buf.WriteString(fmt.Sprintf("\n## %s\n\n", tag.ShowName()))
	buf.WriteString("| # | 题名 | 标签 | 难度 |\n")
	buf.WriteString("| :-: | - | - | :-: |\n")
	format := "| %s | [%s](https://github.com/openset/leetcode/tree/master/problems/%s) | %s | %s |\n"
	maxId := 0
	rows := make(map[int]string)
	for _, question := range questions {
		id, err := strconv.Atoi(question.QuestionFrontendId)
		checkErr(err)
		if question.TranslatedTitle == "" {
			question.TranslatedTitle = question.Title
		}
		rows[id] = fmt.Sprintf(format, question.QuestionFrontendId, question.TranslatedTitle, question.TitleSlug, question.TagsStr(), question.Difficulty)
		if id > maxId {
			maxId = id
		}
	}
	for i := maxId; i > 0; i-- {
		if row, ok := rows[i]; ok {
			buf.WriteString(row)
		}
	}
	filename := path.Join("tag", tag.Slug, "README.md")
	filePutContents(filename, buf.Bytes())
}

func (tag tagType) ShowName() string {
	if tag.TranslatedName != "" {
		return tag.TranslatedName
	}
	return tag.Name
}
