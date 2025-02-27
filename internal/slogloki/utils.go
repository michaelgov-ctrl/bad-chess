package slogloki

import (
	"context"
	"log/slog"
	"reflect"
	"runtime"
	"slices"
)

type ReplaceAttrFn = func(groups []string, a slog.Attr) slog.Attr

func AppendAttrsToGroup(groups []string, actualAttrs []slog.Attr, newAttrs ...slog.Attr) []slog.Attr {
	if len(groups) == 0 {
		actualAttrsCopy := make([]slog.Attr, 0, len(actualAttrs)+len(newAttrs))
		actualAttrsCopy = append(actualAttrsCopy, actualAttrs...)
		actualAttrsCopy = append(actualAttrsCopy, newAttrs...)
		return UniqAttrs(actualAttrsCopy)
	}

	actualAttrs = slices.Clone(actualAttrs)

	for i := range actualAttrs {
		attr := actualAttrs[i]
		if attr.Key == groups[0] && attr.Value.Kind() == slog.KindGroup {
			actualAttrs[i] = slog.Group(groups[0], ToAnySlice(AppendAttrsToGroup(groups[1:], attr.Value.Group(), newAttrs...))...)
			return actualAttrs
		}
	}

	return UniqAttrs(
		append(
			actualAttrs,
			slog.Group(
				groups[0],
				ToAnySlice(AppendAttrsToGroup(groups[1:], []slog.Attr{}, newAttrs...))...,
			),
		),
	)
}

func AppendRecordAttrsToAttrs(attrs []slog.Attr, groups []string, record *slog.Record) []slog.Attr {
	output := make([]slog.Attr, 0, len(attrs)+record.NumAttrs())
	output = append(output, attrs...)

	record.Attrs(func(attr slog.Attr) bool {
		for i := len(groups) - 1; i >= 0; i-- {
			attr = slog.Group(groups[i], attr)
		}
		output = append(output, attr)
		return true
	})

	return output
}

func AttrsToMap(attrs ...slog.Attr) map[string]any {
	output := map[string]any{}

	attrsByKey := groupValuesByKey(attrs)
	for k, values := range attrsByKey {
		v := mergeAttrValues(values...)
		if v.Kind() == slog.KindGroup {
			output[k] = AttrsToMap(v.Group()...)
		} else {
			output[k] = v.Any()
		}
	}

	return output
}

func ContextExtractor(ctx context.Context, fns []func(ctx context.Context) []slog.Attr) []slog.Attr {
	attrs := []slog.Attr{}
	for _, fn := range fns {
		attrs = append(attrs, fn(ctx)...)
	}
	return attrs
}

func FilterMap[T any, R any](collection []T, callback func(item T, index int) (R, bool)) []R {
	result := []R{}

	for i := range collection {
		if r, ok := callback(collection[i], i); ok {
			result = append(result, r)
		}
	}

	return result
}

func FormatError(err error) map[string]any {
	return map[string]any{
		"kind":  reflect.TypeOf(err).String(),
		"error": err.Error(),
		"stack": nil, // @TODO
	}
}

func groupValuesByKey(attrs []slog.Attr) map[string][]slog.Value {
	result := map[string][]slog.Value{}

	for _, item := range attrs {
		key := item.Key
		result[key] = append(result[key], item.Value)
	}

	return result
}

func mergeAttrValues(values ...slog.Value) slog.Value {
	v := values[0]

	for i := 1; i < len(values); i++ {
		if v.Kind() != slog.KindGroup || values[i].Kind() != slog.KindGroup {
			v = values[i]
			continue
		}

		v = slog.GroupValue(append(v.Group(), values[i].Group()...)...)
	}

	return v
}

func RemoveEmptyAttrs(attrs []slog.Attr) []slog.Attr {
	return FilterMap(attrs, func(attr slog.Attr, _ int) (slog.Attr, bool) {
		if attr.Key == "" {
			return attr, false
		}

		if attr.Value.Kind() == slog.KindGroup {
			values := RemoveEmptyAttrs(attr.Value.Group())
			if len(values) == 0 {
				return attr, false
			}

			attr.Value = slog.GroupValue(values...)
			return attr, true
		}

		return attr, !attr.Value.Equal(slog.Value{})
	})
}

func ReplaceAttrs(fn ReplaceAttrFn, groups []string, attrs ...slog.Attr) []slog.Attr {
	for i := range attrs {
		attr := attrs[i]
		value := attr.Value.Resolve()
		if value.Kind() == slog.KindGroup {
			attrs[i].Value = slog.GroupValue(ReplaceAttrs(fn, append(groups, attr.Key), value.Group()...)...)
		} else if fn != nil {
			attrs[i] = fn(groups, attr)
		}
	}

	return attrs
}

func ReplaceError(attrs []slog.Attr, errorKeys ...string) []slog.Attr {
	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if len(groups) > 1 {
			return a
		}

		for i := range errorKeys {
			if a.Key == errorKeys[i] {
				if err, ok := a.Value.Any().(error); ok {
					return slog.Any(a.Key, FormatError(err))
				}
			}
		}
		return a
	}
	return ReplaceAttrs(replaceAttr, []string{}, attrs...)
}

func Source(sourceKey string, r *slog.Record) slog.Attr {
	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	var args []any
	if f.Function != "" {
		args = append(args, slog.String("function", f.Function))
	}
	if f.File != "" {
		args = append(args, slog.String("file", f.File))
	}
	if f.Line != 0 {
		args = append(args, slog.Int("line", f.Line))
	}

	return slog.Group(sourceKey, args...)
}

func ToAnySlice[T any](collection []T) []any {
	result := make([]any, len(collection))
	for i := range collection {
		result[i] = collection[i]
	}
	return result
}

func UniqAttrs(attrs []slog.Attr) []slog.Attr {
	return uniqByLast(attrs, func(item slog.Attr) string {
		return item.Key
	})
}

func uniqByLast[T any, U comparable](collection []T, iteratee func(item T) U) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[U]int, len(collection))
	seenIndex := 0

	for _, item := range collection {
		key := iteratee(item)

		if index, ok := seen[key]; ok {
			result[index] = item
			continue
		}

		seen[key] = seenIndex
		seenIndex++
		result = append(result, item)
	}

	return result
}
