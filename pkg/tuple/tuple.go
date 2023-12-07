package tuple

import (
	"errors"
	"fmt"
	"strings"

	v1 "github.com/Permify/permify-go/generated/base/v1"

	"golang.org/x/exp/slices"
)

const (
	ENTITY    = "%s:%s" // format string for entity in the form of "<type>:<id>"
	RELATION  = "#%s"   // format string for relation in the form of "#<relation>"
	REFERENCE = "%s#%s" // format string for reference in the form of "<type>#<relation>"
)

const (
	ELLIPSIS = "..." // ellipsis string
)

const (
	SEPARATOR = "." // separator string used to concatenate entity and relation
)

// IsDirectSubject checks if the given subject is of type "user"
func IsDirectSubject(subject *v1.Subject) bool {
	return subject.GetRelation() == ""
}

// NormalizeRelation normalizes the relation, treating ellipsis as an empty string
func NormalizeRelation(relation string) string {
	if relation == ELLIPSIS {
		return ""
	}
	return relation
}

// AreSubjectsEqual checks if two subjects are equal
func AreSubjectsEqual(s1, s2 *v1.Subject) bool {
	return NormalizeRelation(s1.GetRelation()) == NormalizeRelation(s2.GetRelation()) && s1.GetId() == s2.GetId() && s1.GetType() == s2.GetType()
}

// AreQueryAndSubjectEqual checks if a query and a subject are equal
func AreQueryAndSubjectEqual(en *v1.Entity, permission string, s2 *v1.Subject) bool {
	return NormalizeRelation(permission) == NormalizeRelation(s2.GetRelation()) && en.GetId() == s2.GetId() && en.GetType() == s2.GetType()
}

// EAREqual checks if two subjects are equal
func EAREqual(s1, s2 *v1.EntityAndRelation) bool {
	return s1.GetRelation() == s2.GetRelation() && s1.GetEntity().GetId() == s2.GetEntity().GetId() && s1.GetEntity().GetType() == s2.GetEntity().GetType()
}

// SubjectToEAR converts a Subject object to an EntityAndRelation object
func SubjectToEAR(subject *v1.Subject) *v1.EntityAndRelation {
	return &v1.EntityAndRelation{
		Entity:   &v1.Entity{Id: subject.GetId(), Type: subject.GetType()},
		Relation: subject.GetRelation(),
	}
}

// EntityAndRelationToString converts an EntityAndRelation object to string format
func EntityAndRelationToString(entity *v1.Entity, relation string) string {
	return EntityToString(entity) + fmt.Sprintf(RELATION, relation)
}

// EntityToString converts an Entity object to string format
func EntityToString(entity *v1.Entity) string {
	return fmt.Sprintf(ENTITY, entity.GetType(), entity.GetId())
}

// SubjectToString converts a Subject object to string format.
func SubjectToString(subject *v1.Subject) string {
	// Convert the subject's type and id to a string in the format of an entity
	entity := fmt.Sprintf(ENTITY, subject.GetType(), subject.GetId())

	// If the subject is a user, return the entity string
	if IsDirectSubject(subject) {
		return entity
	}

	// If the subject is not a user, add the relation to the string
	return fmt.Sprintf("%s"+RELATION, entity, subject.GetRelation())
}

// ToString function converts a Tuple object to a string format.
func ToString(tup *v1.Tuple) string {
	// Retrieve the individual elements of the tuple
	entity := tup.GetEntity()
	relation := tup.GetRelation()
	subject := tup.GetSubject()

	// Convert the elements to strings
	strEntity := EntityToString(entity)
	strRelation := relation
	strSubject := SubjectToString(subject)

	// Combine the strings with proper formatting
	result := fmt.Sprintf("%s#%s@%s", strEntity, strRelation, strSubject)

	// Return the formatted string
	return result
}

// IsEntityAndSubjectEquals checks if the entity and subject of a Tuple object are equal
func IsEntityAndSubjectEquals(t *v1.Tuple) bool {
	return t.GetEntity().GetType() == t.GetSubject().GetType() && t.GetEntity().GetId() == t.GetSubject().GetId() && t.GetRelation() == t.GetSubject().GetRelation()
}

// ValidateSubjectType validates if the subject type and relation are present in the list of allowed relation types
func ValidateSubjectType(subject *v1.Subject, relationTypes []string) (err error) {
	if len(relationTypes) == 0 {
		return errors.New(v1.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String())
	}

	key := subject.GetType()
	if subject.GetRelation() != "" && subject.GetRelation() != ELLIPSIS {
		key += "#" + subject.GetRelation() // append relation to key
	}

	if !slices.Contains(relationTypes, key) { // check if key is in relationTypes
		return errors.New(v1.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String()) // return error if not found
	}
	return nil // return nil if validation succeeds
}

// SplitRelation splits a relation string by the separator "." and returns the result as a slice
func SplitRelation(relation string) (a []string) {
	s := strings.Split(relation, SEPARATOR) // split relation by the separator "."
	a = append(a, s...)
	if len(a) == 1 {
		a = append(a, "") // if there is only one element in the slice, add an empty string to the end
	}
	return
}

// IsRelationComputed checks if a relation is computed or not
func IsRelationComputed(relation string) bool {
	sp := strings.Split(relation, SEPARATOR)
	return len(sp) == 1
}

// IsSubjectValid checks if a subject is valid or not
func IsSubjectValid(subject *v1.Subject) bool {
	if subject.GetType() == "" {
		return false
	}

	if subject.GetId() == "" {
		return false
	}

	if IsDirectSubject(subject) {
		return subject.GetRelation() == "" // relation should be empty for user subjects
	}
	return subject.GetRelation() != "" // relation should not be empty for non-user subjects
}

// Tuple parses a tuple string and returns a Tuple object
func Tuple(tuple string) (*v1.Tuple, error) {
	s := strings.Split(strings.TrimSpace(tuple), "@") // split tuple string by "@"
	if len(s) != 2 {
		return nil, ErrInvalidTuple // return error if number of "@" is not equal to 2
	}
	ear, err := EAR(s[0]) // parse entity and relation from the first part of the tuple string
	if err != nil {
		return nil, err
	}
	sub, err := EAR(s[1]) // parse entity and relation from the second part of the tuple string
	if err != nil {
		return nil, err
	}
	return &v1.Tuple{
		Entity:   ear.Entity,
		Relation: ear.Relation,
		Subject: &v1.Subject{
			Type:     sub.Entity.Type,
			Id:       sub.Entity.Id,
			Relation: sub.Relation,
		},
	}, nil
}

// EAR function parses a string to create a v1.EntityAndRelation object.
func EAR(ear string) (*v1.EntityAndRelation, error) {
	// Split EntityAndRelation string by "#" and trim spaces
	s := strings.Split(strings.TrimSpace(ear), "#")

	// Check if there is at least one part (entity) in the string
	if len(s) < 1 {
		return nil, ErrInvalidEntityAndRelation
	}

	// Parse entity from the first part of the string
	e, err := E(s[0])
	if err != nil {
		return nil, err
	}

	// Create a new EntityAndRelation with the parsed entity
	entityAndRelation := &v1.EntityAndRelation{
		Entity: e,
	}

	// If there is a second part (relation), add it to EntityAndRelation
	if len(s) > 1 {
		entityAndRelation.Relation = s[1]
	}

	// Return the created EntityAndRelation
	return entityAndRelation, nil
}

// E function parses an Entity string and returns an Entity object.
func E(e string) (*v1.Entity, error) {
	// Split Entity string by ":" and trim spaces
	s := strings.Split(strings.TrimSpace(e), ":")

	// Check if the string has exactly two parts (Type and Id)
	if len(s) != 2 {
		return nil, ErrInvalidEntity // Return error if number of ":" is not exactly 2
	}

	// Check if Type and Id are not empty
	if s[0] == "" || s[1] == "" {
		return nil, ErrInvalidEntity // Return error if either Type or Id is empty
	}

	// Return the created Entity
	return &v1.Entity{
		Type: s[0],
		Id:   s[1],
	}, nil
}

// ReferenceToString -
func ReferenceToString(ref *v1.RelationReference) string {
	if ref.GetRelation() != "" {
		return fmt.Sprintf(REFERENCE, ref.GetType(), ref.GetRelation())
	}
	return ref.GetType()
}

// RelationReference parses a relation reference string and returns a RelationReference object.
func RelationReference(ref string) *v1.RelationReference {
	// Split the reference string by "#"
	sp := strings.Split(ref, "#")

	// Create a new RelationReference with the parsed Type
	relationReference := &v1.RelationReference{
		Type: sp[0],
	}

	// If there is a second part (Relation), add it to RelationReference
	if len(sp) > 1 {
		relationReference.Relation = sp[1]
	}

	// Return the created RelationReference or an error if any step failed
	return relationReference
}

// AreRelationReferencesEqual checks if two relation references are equal or not
func AreRelationReferencesEqual(s1, s2 *v1.RelationReference) bool {
	return s1.GetRelation() == s2.GetRelation() && s1.GetType() == s2.GetType()
}
