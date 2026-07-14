import { EditorState } from "@codemirror/state";
import { EditorView, basicSetup } from "codemirror";
import { yaml } from "@codemirror/lang-yaml";
import type React from "react";
import { useEffect, useRef } from "react";
import { css } from "@emotion/css";

const editorStyles = {
  container: css`
    border: 1px solid #d2d2d2;
    border-radius: 4px;
    overflow: hidden;
    font-size: 13px;

    .cm-editor {
      height: 100%;
    }
    .cm-scroller {
      min-height: 300px;
      max-height: 600px;
    }
    .cm-gutters {
      background-color: #f5f5f5;
      border-right: 1px solid #d2d2d2;
    }
  `,
};

interface YamlEditorProps {
  value: string;
  onChange: (value: string) => void;
  readOnly?: boolean;
}

export const YamlEditor: React.FC<YamlEditorProps> = ({
  value,
  onChange,
  readOnly = false,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    if (!containerRef.current) return;

    const state = EditorState.create({
      doc: value,
      extensions: [
        basicSetup,
        yaml(),
        EditorState.readOnly.of(readOnly),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            onChangeRef.current(update.state.doc.toString());
          }
        }),
      ],
    });

    const view = new EditorView({
      state,
      parent: containerRef.current,
    });

    viewRef.current = view;

    return () => {
      view.destroy();
    };
  }, [readOnly]);

  useEffect(() => {
    const view = viewRef.current;
    if (!view) return;
    const current = view.state.doc.toString();
    if (current !== value) {
      view.dispatch({
        changes: { from: 0, to: current.length, insert: value },
      });
    }
  }, [value]);

  return <div ref={containerRef} className={editorStyles.container} />;
};
